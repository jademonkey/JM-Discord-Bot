// Start
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
)

func init() {
	flag.StringVar(&token, "t", "", "Bot Token")
	flag.StringVar(&channelMonitor, "c", "", "Channel ID to monitor")
	flag.Parse()
}

var token string

var channelMonitor string

func main() {
	log.Printf("Starting up")

	errorNow := false

	// TODO: READ files for secrets instead of commandline!
	if token == "" {
		fmt.Println("No token provided. Please run: murderbot -t <bot token>")
		errorNow = true
	}

	if channelMonitor == "" {
		fmt.Println("No channelID provided. Please run: muderbot -c <channelID>")
		errorNow = true
	}

	if errorNow {
		return
	}

	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		fmt.Println("Error creating Discord session: ", err)
		return
	}

	// Register ready as a callback for the ready events.
	dg.AddHandler(ready)

	dg.AddHandler(messageCreate)

	// Open a websocket connection to Discord and begin listening.
	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
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
	if m.ChannelID != channelMonitor {
		log.Printf("Channel %s is not the right channel %s! (%s:%s)", m.ChannelID, channelMonitor, m.Author, m.Content)
		return
	}

	log.Printf("I saw the following message from %s in the %s channel: %s", m.Author, m.ChannelID, m.Content)
}
