package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"
	"github.com/bwmarrin/discordgo"
)

var fortunes []string
var offensive []string

/* Scanner split function, reads in file and delimits by % */
func FortuneSplit(data []byte, atEOF bool) (advance int, token []byte, err error) {
	for i := 0; i < len(data); i++ {
		if data[i] == '%' {
			return i + 1, data[:i], nil
		}
	}
	if !atEOF {
		return 0, nil, nil
	}
	return 0, data, bufio.ErrFinalToken
}

/* Runs through the given file using the splitter function and fills the fortune slices */
func ParseFortune(f *os.File, o bool) {
	scanner := bufio.NewScanner(f)
	scanner.Split(FortuneSplit)
	for scanner.Scan() {
		if o {
			offensive = append(fortunes, scanner.Text())
		} else {
			fortunes = append(fortunes, scanner.Text())
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "reading input:%v\n", err)
	}
}

/* Return a fortune randomly */
func GetFortune() (fortune string) {
	i := rand.Intn(len(fortunes))
	return fortunes[i]
}

/* Upon receiving a command, bot sends a fortune */
func SendFortune(s *discordgo.Session, m *discordgo.MessageCreate) {
	// ignore bot messages
	if m.Author.ID == s.State.User.ID {
		return
	}

	if m.Content == "!fortune" {
		msgtitle := m.Author.Username + ", it's your lucky day."
		me := discordgo.MessageEmbed{ Title: msgtitle, Description: GetFortune(), Color: 39423 }
		s.ChannelMessageSendEmbed(m.ChannelID, &me)
	} else if m.Content == "!offendme" {
		msgtitle := m.Author.Username + ", it's your lucky day."
		me := discordgo.MessageEmbed{ Title: msgtitle, Description: GetFortune(), Color: 14103594 }
		s.ChannelMessageSendEmbed(m.ChannelID, &me)
	}
}

func main() {
	fmt.Printf("IT'S YOUR LUCKY DAY\n")

	f, err := os.Open("fortunes")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading fortune file: %v\n", err)
		os.Exit(1)
	}
	ParseFortune(f, false)
	f.Close()

	f, err = os.Open("offensive")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading offensive fortune file: %v\n", err)
		os.Exit(1)
	}
	ParseFortune(f, true)
	f.Close()

	if len(os.Args) != 2 {
		fmt.Printf("Usage: zoltar [token]\n")
		os.Exit(0)
	}

	var token string = os.Args[1]
	rand.Seed(time.Now().UTC().UnixNano())

	// Create Discord session
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating Discord session: %v\n", err)
		os.Exit(1)
	}

	dg.AddHandler(SendFortune)
	dg.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsGuildMessages)

	// Open websocket and begin listening
	err = dg.Open()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening Discord session: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Listening for commands...\n")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	fmt.Printf("\nInterrupt Received! Exiting.\n")
	dg.Close()
}
