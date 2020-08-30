// (c) 2020 Robert Parker
// This code is licensed under MIT license (see LICENSE for details)
package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
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
	// Cleanly close down the Discord session.
	dg.Close()
}

// This function will be called (due to AddHandler above) when the bot receives
// the "ready" event from Discord.
func ready(s *discordgo.Session, event *discordgo.Ready) {

	log.Println("Discord has alerted us it is ready for us!")

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
