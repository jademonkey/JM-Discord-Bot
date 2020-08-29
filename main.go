// Start
package main

import (
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
)

var token string
var channel string
var guild string

//temp
var admins string
var rolesStrings string
var bannedusers string

func readFiles() {
	// token
	tempBytes, err := ioutil.ReadFile("data/token")
	if err != nil {
		log.Printf("failed to read data/token file.")
		return
	}
	token = string(tempBytes)

	// channel
	tempBytes, err = ioutil.ReadFile("data/channel")
	if err != nil {
		log.Printf("failed to read data/channel file.")
		return
	}
	channel = string(tempBytes)

	// guild
	tempBytes, err = ioutil.ReadFile("data/guild")
	if err != nil {
		log.Printf("failed to read data/channel file.")
		return
	}
	guild = string(tempBytes)

	//admins
	tempBytes, err = ioutil.ReadFile("data/admins")
	if err != nil {
		log.Printf("failed to read data/admins file.")
		return
	}
	admins = string(tempBytes)

	// roles
	tempBytes, err = ioutil.ReadFile("data/roles")
	if err != nil {
		log.Printf("failed to read data/admins file.")
		return
	}
	rolesStrings = string(tempBytes)

	// banned users
	tempBytes, err = ioutil.ReadFile("data/bans")
	if err != nil {
		log.Printf("failed to read data/admins file.")
		return
	}
	bannedusers = string(tempBytes)

	// temp output
	log.Printf("token read '%s'", token)
	log.Printf("channel read '%s'", channel)
	log.Printf("guild read '%s'", guild)
	log.Printf("admins read '%s'", admins)
	log.Printf("roles read '%s'", rolesStrings)
	log.Printf("bans read '%s'", bannedusers)
}

// Format: <friendly print name> <user ID>
func parseAdmins() {

}

// Format: <friendly print name> <user ID>
func parseBans() {

}

// Format: <string variant> <role ID>
func parseRoles() {

}

func main() {
	log.Printf("Starting up")

	readFiles()

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

	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Println("Error creating Discord session: ", err)
		return
	}

	// Register ready as a callback for the ready events.
	dg.AddHandler(ready)

	dg.AddHandler(messageCreate)

	// Open a websocket connection to Discord and begin listening.
	err = dg.Open()
	if err != nil {
		log.Println("error opening connection,", err)
		return
	}

	// Wait here until CTRL-C or other term signal is received.
	log.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session.
	dg.Close()
}

// This function will be called (due to AddHandler above) when the bot receives
// the "ready" event from Discord.
func ready(s *discordgo.Session, event *discordgo.Ready) {

	// Set the playing status.
	log.Printf("Received Ready Event")
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
