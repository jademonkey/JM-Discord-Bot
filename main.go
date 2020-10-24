// (c) 2020 Robert Parker
// This code is licensed under MIT license (see LICENSE for details)
package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
)

// Data types
type fileEntry struct {
	FriendlyName string
	ID           string
}

// Constants
const adminFile = "data/admins"
const bansFile = "data/bans"
const channelFile = "data/channel"
const guildFile = "data/guild"
const rolesFile = "data/roles"
const tokenFile = "data/token"

// Global variables
var token string
var channel string
var guild string
var admins []fileEntry
var rolesStrings []fileEntry
var bannedusers []fileEntry

var usagePrint map[string]string
var commandFuncs map[string]func(s *discordgo.Session, m *discordgo.MessageCreate, parms []string)

func populateUsage() {
	usagePrint = make(map[string]string)

	usagePrint["help"] = "!help (command) - Print available commands or more details about a single command"
	usagePrint["roll"] = "!roll <dice>    - Rolls a specific dice. Format is xdx, where x is a positive number."
	usagePrint["listmyroles"] = "!listmyroles - Prints your current assigned roles."
	usagePrint["addrole"] = "!addrole <role> - Adds the requested role to you."
	usagePrint["removerole"] = "!removerole <role> - Reoves the role from you."
	usagePrint["listroles"] = "!listroles - Lists roles you can assign to yourself."
}

func populateFuncs() {
	commandFuncs = make(map[string]func(s *discordgo.Session, m *discordgo.MessageCreate, parms []string))

	commandFuncs["help"] = callHelp
	commandFuncs["roll"] = callRoll
	commandFuncs["listmyroles"] = callListMyRoles
	commandFuncs["addrole"] = callAddRole
	commandFuncs["removerole"] = callRemoveRole
	commandFuncs["listroles"] = callListRoles
}

// Function to read necessary files and load the global
// variables with the data in them
func readDataFiles() error {
	// token
	var err error
	token, err = readSingleLineFile(tokenFile)
	if err != nil {
		log.Printf("failed to read %s file. %v", tokenFile, err)
		return err
	}

	// channel
	channel, err = readSingleLineFile(channelFile)
	if err != nil {
		log.Printf("failed to read %s file. %v", channelFile, err)
		return err
	}

	// guild
	guild, err = readSingleLineFile(guildFile)
	if err != nil {
		log.Printf("failed to read %s file. %v", guildFile, err)
		return err
	}

	//admins
	admins, err = readEntriesFromFile(adminFile)
	if err != nil {
		log.Printf("failed to read %s file. %v", adminFile, err)
		return err
	}

	// roles
	rolesStrings, err = readEntriesFromFile("data/roles")
	if err != nil {
		log.Printf("failed to read data/roles file. %v", err)
		return err
	}

	// banned users
	bannedusers, err = readEntriesFromFile("data/bans")
	if err != nil {
		log.Printf("failed to read data/bans file. %v", err)
		return err

	}

	// temp output
	log.Printf("token read '%v'", token)
	log.Printf("channel read '%v'", channel)
	log.Printf("guild read '%v'", guild)
	log.Printf("admins read '%v'", admins)
	log.Printf("roles read '%v'", rolesStrings)
	log.Printf("bans read '%v'", bannedusers)
	return nil
}

// Main execution function
func main() {
	log.Println("Starting up")

	err := readDataFiles()
	if err != nil {
		log.Fatalf("Error reading data files: %v", err)
	}

	// Validation
	if token == "" {
		log.Fatalln("Token file was read but could not extract the token!")
	}
	if channel == "" {
		log.Fatalln("Channel file was read but could not extract the channel ID!")
	}
	if guild == "" {
		log.Fatalln("Guild ID file was read but could not extract the guild ID!")
	}

	log.Println("Data Files read")

	populateUsage()
	populateFuncs()

	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Fatalln("Error creating Discord session: ", err)
	}

	// Register ready as a callback for the ready events.
	dg.AddHandler(ready)

	dg.AddHandler(messageCreate)

	log.Println("Connecting Bot")

	// Open a websocket connection to Discord and begin listening.
	err = dg.Open()
	if err != nil {
		log.Fatalln("error opening connection,", err)
	}

	dg.UpdateStatus(0, "Listening to !help")

	// Wait here until CTRL-C or other term signal is received.
	log.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	log.Println("Closing down")
	//dg.ChannelMessageSend(channel, "I am shutting down, i will not be able to handle your messages!")
	// Cleanly close down the Discord session.
	dg.Close()
}

// This function will be called (due to AddHandler above) when the bot receives
// the "ready" event from Discord.
func ready(s *discordgo.Session, event *discordgo.Ready) {
	log.Println("Discord has alerted us it is ready for us!")
	//s.ChannelMessageSend(channel, "I am Ready for action!")
}

// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the authenticated bot has access to.
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {

	// Ignore all messages created by the bot itself
	// This isn't required in this specific example but it's a good practice.
	if m.Author.ID == s.State.User.ID {
		return
	}

	// Ignore all messages not in the specific bot channel
	if m.ChannelID != channel {
		log.Printf("Channel %s is not the right channel %s! (%s:%s)", m.ChannelID, channel, m.Author, m.Content)
		return
	}

	log.Printf("I saw the following message from %s in the %s channel: %s", m.Author, m.ChannelID, m.Content)

	if strings.HasPrefix(m.Content, "!") {
		// It's a command!
		log.Println("Detected a command!")
		allParms := strings.Split(m.Content, " ")
		command := strings.TrimPrefix(allParms[0], "!")
		callCommand(s, m, command, allParms[1:len(allParms)])
	}
}

// Reads a single line from a given file or returns an error
// if it cannot read the file or there is more than 1 line.
func readSingleLineFile(file string) (string, error) {
	tempBytes, err := ioutil.ReadFile(file)
	if err != nil {
		log.Printf("failed to read %s file. %v", file, err)
		return "", err
	}
	strTempBytes := strings.Trim(string(tempBytes), " ")
	if strings.Contains(strTempBytes, "\n") {
		return "", fmt.Errorf("File contained more than 1 line")
	}

	return strTempBytes, nil
}

// Reads a file for Entries and returns an array of those entries
// All entries are in the format '<string> <string>'
func readEntriesFromFile(file string) ([]fileEntry, error) {
	var toreturn []fileEntry
	fileHandle, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer fileHandle.Close()

	lineReader := bufio.NewReader(fileHandle)
	for {
		line, err := lineReader.ReadString('\n')
		if err != nil && err != io.EOF {
			log.Printf("error - %v", err)
			return nil, err
		}
		line = strings.Trim(line, " \n\r\t")
		if len(line) == 0 && err == nil {
			continue
		}
		if len(line) == 0 && err != nil {
			break
		}

		sepd := strings.Split(line, " ")
		if len(sepd) != 2 {
			// Faulty line so ignore
			continue
		}

		toAdd := fileEntry{sepd[0], sepd[1]}
		toreturn = append(toreturn, toAdd)

		if err != nil {
			break
		}
	}

	return toreturn, nil
}

// Takes an array of entries and writes them to a given file
func writeEntriesToFile(file string, entries []fileEntry) error {
	// TODO
	return nil
}

func callCommand(s *discordgo.Session, m *discordgo.MessageCreate, command string, parms []string) {
	log.Printf("Command called '%s' with parameters '%s'\n", command, parms)

	if val, ok := commandFuncs[command]; ok {
		val(s, m, parms)
	} else {
		printUsage(s, m, "")
	}
}

func callListMyRoles(s *discordgo.Session, m *discordgo.MessageCreate, parms []string) {
	//parms ignored
	userID := m.Author.ID
	var toPrint string

	toPrint = m.Author.Mention() + " Your roles are: "

	// Get the user object!
	memberObj, err := s.GuildMember(guild, userID)
	if err != nil {
		log.Printf("Failed to get the member from the guild: %v\n", err)
	}

	//Get the roles list
	allRolesID := memberObj.Roles

	i := 0
	for _, roleI := range allRolesID {
		for _, roleF := range rolesStrings {
			if roleF.ID == roleI {
				if i == 0 {
					i++
				} else {
					toPrint += ", "
				}
				toPrint += "`" + roleF.FriendlyName + "`"
				break
			}
		}
	}

	s.ChannelMessageSend(channel, toPrint)
}

func callAddRole(s *discordgo.Session, m *discordgo.MessageCreate, parms []string) {
}

func callRemoveRole(s *discordgo.Session, m *discordgo.MessageCreate, parms []string) {

}

func callListRoles(s *discordgo.Session, m *discordgo.MessageCreate, parms []string) {
	//parms ignored
	var toPrint string
	log.Println("Listing the available roles")

	toPrint = m.Author.Mention() + " The following roles are available: "
	i := 0
	for _, role := range rolesStrings {
		if i != 0 {
			toPrint += ", "
		} else {
			i++
		}
		toPrint += "`" + role.FriendlyName + "`"
	}

	s.ChannelMessageSend(channel, toPrint)
}

func callHelp(s *discordgo.Session, m *discordgo.MessageCreate, parms []string) {
	if len(parms) == 0 {
		log.Println("help: no parms provided")
		printUsage(s, m, "")
	} else if len(parms) == 1 {
		log.Printf("help: 1 parm provided %s\n", parms)
		printUsage(s, m, parms[0])
	} else {
		log.Println("help: bad call")
		printUsage(s, m, "help")
	}
}

func callRoll(s *discordgo.Session, m *discordgo.MessageCreate, parms []string) {
	if len(parms) != 1 {
		log.Printf("Incorrect number of parameters %d\n", len(parms))
		printUsage(s, m, "roll")
		return
	}

	figments := strings.Split(parms[0], "d")
	if len(figments) != 2 {
		log.Printf("Incorrect number of small parts %d\n", len(figments))
		printUsage(s, m, "roll")
		return
	}

	i1, err := strconv.Atoi(figments[0])
	if err != nil {
		log.Printf("First component is not a number - %v\n", err)
		printUsage(s, m, "roll")
		return
	}
	i2, err := strconv.Atoi(figments[1])
	if err != nil {
		log.Printf("Second component is not a number - %v\n", err)
		printUsage(s, m, "roll")
		return
	}

	if i1 < 0 || i2 < 0 {
		log.Println("One of the number is 0")
		printUsage(s, m, "roll")
		return
	}

	toPrint := m.Author.Mention() + " You rolled " + parms[0] + "\nResult: "

	if i1 == 0 || i2 == 0 {
		toPrint += "0 - dumbass"
	} else {
		max := i1 * i2
		myNum := rand.Intn(max) // In range 0 -> max-1
		myNum++                 //So we are in range 1 -> max
		toPrint += strconv.Itoa(myNum) + "\n"
	}

	s.ChannelMessageSend(channel, toPrint)
}

func printUsage(s *discordgo.Session, m *discordgo.MessageCreate, command string) {
	var toPrint string
	log.Printf("searching for usage for %s\n", command)
	toPrint = m.Author.Mention() + " "
	if command != "" {
		toPrint += usagePrint[command]
		if toPrint == "" {
			log.Println("Printing all usages")
			toPrint += "Unknown command: " + command + "\n"
			for _, value := range usagePrint {
				toPrint += value + "\n"
			}
		}
	} else {
		log.Println("Printing all usages")
		for _, value := range usagePrint {
			toPrint += value + "\n"
		}
	}
	s.ChannelMessageSend(channel, toPrint)
}
