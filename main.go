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

	"strings"

	"strconv"

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
			Fish        string `json:"fish"`
			Location    string `json:"location"`
			Blacklist   string `json:"blacklist"`
			Inventory   string `json:"inventory"`
			Gather      string `json:"gather"`
			Leaderboard string `json:"leaderboard"`
			Time        string `json:"time"`
		} `json:"paths"`
	} `json:"api"`
}

var items = map[string]bool{
	"rod":     true,
	"hook":    true,
	"bait":    true,
	"vehicle": true,
	"baitbox": true,
}

var locations = map[string]bool{
	"lake":  true,
	"river": true,
	"ocean": true,
}

var config Config
var Client = &http.Client{}

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
		err := fish(s, m.Message)
		if err != nil {
			fmt.Println("Error posting to API " + err.Error())
			return
		}
		//s.ChannelMessageSend(m.ChannelID, fish)
	}
	if strings.HasPrefix(m.Content, "pls location") {
		loc := strings.TrimPrefix(m.Content, "pls location ")
		if !locations[loc] {
			s.ChannelMessageSend(m.ChannelID, ":x: | The locations are lake, river, and ocean retard")
			return
		}
		req, err := http.NewRequest("PUT",
			config.API.BaseURL+config.API.Paths.Location+m.Author.ID+"/"+loc,
			nil)
		if err != nil {
			fmt.Println(err)
			return
		}
		s.ChannelMessageSend(m.ChannelID, reqAndGetMsg(req))
		return
	}
	if strings.HasPrefix(m.Content, "pls blacklist") {
		if !Admins[m.Author.ID] {
			s.ChannelMessageSend(m.ChannelID, "You do not have the required permissions")
			return
		}
		req, err := http.NewRequest("GET", config.API.BaseURL+config.API.Paths.Blacklist+strings.TrimPrefix(m.Content, "pls blacklist "), nil)
		if err != nil {
			fmt.Println(err)
			return
		}
		s.ChannelMessageSend(m.ChannelID, reqAndGetMsg(req))
		return
	}
	if strings.HasPrefix(m.Content, "pls unblacklist") {
		if !Admins[m.Author.ID] {
			s.ChannelMessageSend(m.ChannelID, "You do not have the required permissions")
			return
		}
		req, err := http.NewRequest("DELETE", config.API.BaseURL+config.API.Paths.Blacklist+strings.TrimPrefix(m.Content, "pls unblacklist "), nil)
		if err != nil {
			fmt.Println(err)
			return
		}
		s.ChannelMessageSend(m.ChannelID, reqAndGetMsg(req))
		return
	}
	if strings.HasPrefix(m.Content, "pls buy") {
		toBuy := strings.Split(strings.TrimPrefix(m.Content, "pls buy "), " ")
		if len(toBuy) < 2 {
			s.ChannelMessageSend(m.ChannelID, ":x: | please put the tier of the item you would like to buy <:Kizuna_Blur:311262010157563907>")
			return
		}
		if !items[toBuy[0]] {
			s.ChannelMessageSend(m.ChannelID, ":x: | invalid item >.>")
			return
		}
		if i, _ := strconv.Atoi(toBuy[1]); i > 5 {
			s.ChannelMessageSend(m.ChannelID, ":x: | tiers are between 1 and 5")
			return
		}
		if i, _ := strconv.Atoi(toBuy[1]); i < 1 {
			s.ChannelMessageSend(m.ChannelID, ":x: | tiers are between 1 and 5")
			return
		}

		req, err := http.NewRequest("POST",
			config.API.BaseURL+config.API.Paths.Inventory+m.Author.ID,
			bytes.NewBuffer([]byte(fmt.Sprintf(`{"item": "%v", "tier": "%v"}`, toBuy[0], toBuy[1]))))
		if err != nil {
			fmt.Println(err)
			return
		}
		s.ChannelMessageSend(m.ChannelID, reqAndGetMsg(req))
		return
	}
	if strings.HasPrefix(m.Content, "pls gather bait") {
		req, err := http.NewRequest("POST", config.API.BaseURL+config.API.Paths.Gather+m.Author.ID, nil)
		if err != nil {
			fmt.Println(err)
			s.ChannelMessageSend(m.ChannelID, err.Error())
			return
		}
		s.ChannelMessageSend(m.ChannelID, reqAndGetMsg(req))
		return
	}
	if strings.HasPrefix(m.Content, "pls leaderboard") {
		var d map[string]interface{}
		req, err := http.NewRequest("POST",
			config.API.BaseURL+config.API.Paths.Leaderboard,
			bytes.NewBuffer([]byte(fmt.Sprintf(`{"global": true, "page": 1, "user": "%v"}`, m.Author.ID))))
		if err != nil {
			fmt.Println(err)
			s.ChannelMessageSend(m.ChannelID, err.Error())
			return
		}
		json.Unmarshal([]byte(reqAndGetMsg(req)), &d)

		if d["error"].(bool) {
			s.ChannelMessageSend(m.ChannelID, ":x: | "+d["message"].(string))
			return
		}
		s.ChannelMessageSend(m.ChannelID, d["data"].(string))
		return
	}
	if strings.HasPrefix(m.Content, "pls time") {
		type data struct {
			Error   bool   `json:"error"`
			Message string `json:"message"`
			Data    struct {
				Time    string `json:"time"`
				Morning bool   `json:"morning"`
				Night   bool   `json:"night"`
			} `json:"data"`
		}
		var d data
		var msg string
		req, _ := http.NewRequest("GET", config.API.BaseURL+config.API.Paths.Time, nil)
		json.Unmarshal([]byte(reqAndGetMsg(req)), &d)

		if d.Error {
			s.ChannelMessageSend(m.ChannelID, ":x: | "+d.Message)
			return
		}

		if d.Data.Morning {
			msg = "\nYour Tatsumaki:tm: branded almanac suggests that morning fish bite at this time"
		}
		if d.Data.Night {
			msg = "\nYour Tatsumaki:tm: branded almanac suggests that night fish bite at this time"
		}
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("You look down at your watch and see that it reads %v%v", d.Data.Time, msg))
		return
	}
	if m.Content == "pls trash" {
		type data struct {
			Error   bool   `json:"error"`
			Message string `json:"message"`
			Data    string `json:"data"`
		}
		var d data
		req, _ := http.NewRequest("GET", config.API.BaseURL+"/v1/trash", nil)
		json.Unmarshal([]byte(reqAndGetMsg(req)), &d)
		s.ChannelMessageSend(m.ChannelID, d.Data)
	}
	if m.Content == "pls rfish" {
		var embed *discordgo.MessageEmbed
		type data map[string]interface{}
		var d data
		req, _ := http.NewRequest("GET", config.API.BaseURL+"/v1/rfish", nil)
		json.Unmarshal([]byte(reqAndGetMsg(req)), &d)
		_d, _ := json.Marshal(d["data"].(map[string]interface{}))
		json.Unmarshal(_d, &embed)
		s.ChannelMessageSendEmbed(m.ChannelID, embed)
		// s.ChannelMessageSend(m.ChannelID,
		// 	fmt.Sprintf("You caught a tier %s %s. It is %scm long and is worth %s.\n%s", strconv.FormatFloat(d["tier"].(float64), 'f', -1, 32), d["name"], strconv.FormatFloat(d["size"].(float64), 'f', -1, 32), strconv.FormatFloat(d["price"].(float64), 'f', -1, 32), d["pun"]))
	}
	if strings.HasPrefix(m.Content, "pls bait") {
		var d baitInvResp
		args := strings.Split(strings.TrimPrefix(m.Content, "pls bait "), " ")
		if len(args) == 1 {
			if args[0] == "inv" {
				req, _ := http.NewRequest("GET", config.API.BaseURL+"/v1/bait/"+m.Author.ID, nil)
				json.Unmarshal([]byte(reqAndGetMsg(req)), &d)
				s.ChannelMessageSend(m.ChannelID,
					fmt.Sprintf("Tier 1: %v\nTier 2: %v\nTier 3: %v\nTier 4: %v\nTier 5: %v\nTotal bait: %v\nEquipped bait tier: %v\nMax bait carryable: %v", d.Data.Bait.T1, d.Data.Bait.T2, d.Data.Bait.T3, d.Data.Bait.T4, d.Data.Bait.T5, d.Data.CurrentBaitCount, d.Data.CurrentTier, d.Data.MaxBait))
				return
			}
		} else if len(args) == 2 {
			if args[0] == "use" {
				if args[1] <= "5" && args[1] >= "1" {
					req, _ := http.NewRequest("POST", config.API.BaseURL+"/v1/bait/"+m.Author.ID+"/current", bytes.NewBuffer([]byte(fmt.Sprintf(`{"tier": %v}`, args[1]))))
					reqAndGetMsg(req)
					s.ChannelMessageSend(m.ChannelID, "I think it worked")
					return
				}
				s.ChannelMessageSend(m.ChannelID, "tiers are between 1 and 5")
				return
			}
			s.ChannelMessageSend(m.ChannelID, "??? ur doing something wrong")
			return
		} else if len(args) == 3 {
			if args[0] == "buy" {
				if args[1] <= "5" && args[1] >= "1" {
					req, _ := http.NewRequest("POST", config.API.BaseURL+"/v1/bait/"+m.Author.ID+"", bytes.NewBuffer([]byte(fmt.Sprintf(`{"tier": %v, "amount": %v}`, args[1], args[2]))))
					s.ChannelMessageSend(m.ChannelID, reqAndGetMsg(req))
					return
				}
				s.ChannelMessageSend(m.ChannelID, "hey you. use a number between 1 and 5")
				return
			}
			s.ChannelMessageSend(m.ChannelID, "??? uhh ur doing something wrong")
			return
		}
		s.ChannelMessageSend(m.ChannelID, "??? uhh ur doing something super wrong")
		return
	}
}

type baitInvResp struct {
	Error   bool   `json:"error"`
	Message string `json:"message"`
	Data    struct {
		Bait struct {
			T1 int `json:"t1"`
			T2 int `json:"t2"`
			T3 int `json:"t3"`
			T4 int `json:"t4"`
			T5 int `json:"t5"`
		} `json:"bait"`
		CurrentBaitCount int `json:"currentBaitCount"`
		CurrentTier      int `json:"currentTier"`
		MaxBait          int `json:"maxBait"`
	} `json:"data"`
}

// type resp struct {
// 	Err bool,
// 	Message string,
// 	Data string
// }

func reqAndGetMsg(req *http.Request) string {
	resp, err := Client.Do(req)
	if err != nil {
		fmt.Println(err)
		return "error check log"
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return "error check log"
	}
	return string(data)
}

func fish(s *discordgo.Session, msg *discordgo.Message) error {
	type data struct {
		Error   bool                    `json:"error"`
		Message string                  `json:"message"`
		Data    *discordgo.MessageEmbed `json:"data"`
	}
	var r data
	j, _ := json.Marshal(msg)
	resp, err := http.Post(config.API.BaseURL+config.API.Paths.Fish, "application/json", bytes.NewBuffer(j))
	if err != nil {
		return err
	}
	d, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	json.Unmarshal(d, &r)
	if r.Error {
		s.ChannelMessageSend(msg.ChannelID, r.Message)
		return nil
	}
	s.ChannelMessageSendEmbed(msg.ChannelID, r.Data)
	return nil
}
