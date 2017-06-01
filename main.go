package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
)

// Config holds the structure for config.json
type Config struct {
	Bot struct {
		Token  string   `json:"token"`
		Admins []string `json:"admins"`
	} `json:"bot"`
	API struct {
		BaseURL string `json:"baseurl"`
		Paths   struct {
			Fish string `json:"fish"`
		} `json:"paths"`
	} `json:"api"`
}

var config Config

func init() {
	configData, err := ioutil.ReadFile("config.json")
	if err != nil {
		log.Panic("Config not detected in current directory, " + err.Error())
	}

	if err := json.Unmarshal(configData, &config); err != nil {
		log.Panic("Could not unmarshal config, " + err.Error())
	}

}

func main() {
	dg, err := discordgo.New("Bot " + config.Bot.Token)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	dg.AddHandler(messageCreate)

	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	dg.Close()
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}
	if m.Content == "pls fish" {
		fish, err := fish(m.Message)
		if err != nil {
			fmt.Println("Error posting to API " + err.Error())
			return
		}
		s.ChannelMessageSend(m.ChannelID, fish)
	}
}

func fish(msg *discordgo.Message) (string, error) {
	json, _ := json.Marshal(msg)
	resp, err := http.Post(config.API.BaseURL+config.API.Paths.Fish, "application/json", bytes.NewBuffer(json))
	if err != nil {
		return "", err
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
