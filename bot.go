package main

import (
	"bufio"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	// Uncomment below line if you are going to use uptimerobot to ping
	//"net/http"
	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

type UserPowerData struct {
	user_id  string
	guild_id string
	power    int
}

var PowerStructs []UserPowerData

func main() {
	// Uncomment this code block if you are going to use uptimerobot to ping
	// http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
	// 	fmt.Fprintf(w, "Hello, World!")
	// })

	go http.ListenAndServe(":8080", nil)

	file, _ := os.Open("powerdb.txt")
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		tokens := strings.Split(scanner.Text(), " ")
		power_convert, _ := strconv.Atoi(tokens[2])
		current_power_data := UserPowerData{user_id: tokens[0], guild_id: tokens[1], power: power_convert}
		PowerStructs = append(PowerStructs, current_power_data)
	}

	file.Close()

	file.WriteString("b")
	// Load bot token
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	bottoken := os.Getenv("TOKEN")

	// Create a new Discord session using the bot token from .env
	bot, err := discordgo.New("Bot " + bottoken)
	if err != nil {
		panic(err)
	}

	// Register events
	bot.AddHandler(ready)
	bot.AddHandler(messageCreate)

	// Start sesson
	err = bot.Open()
	if err != nil {
		fmt.Println("Error opening Discord session: ", err)
	}

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, syscall.SIGTERM)
	<-sc

	// Cleanly close down the Discord session.
	file_write, _ := os.Create("powerdb.txt")
	for _, item := range PowerStructs {
		kneaded_string := item.user_id + " " + item.guild_id + " " + strconv.Itoa(item.power) + "\n"
		file_write.WriteString(kneaded_string)
	}
	file_write.Close()

	bot.Close()
}

func ready(s *discordgo.Session, event *discordgo.Ready) {
	s.UpdateListeningStatus("your heartbeat")
	fmt.Println("logged in as user " + string(s.State.User.ID))
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.Bot {
		return
	}

	rando := rand.Intn(3-1) + 1

	if (m.Author.ID == "1109446351630123161") && rando <= 1 {
		s.MessageReactionAdd(m.ChannelID, m.ID, "🤓")
	}

	tokens := strings.Split(m.Content, " ")
	skip := true

	for i, item := range PowerStructs {
		if item.guild_id == m.GuildID && item.user_id == m.Author.ID {
			if tokens[0] != "/power" && tokens[0] != "/givepower" {
				PowerStructs[i].power++
			}
			skip = false
			break
		}
	}

	if skip {
		current_power_data := UserPowerData{user_id: m.Author.ID, guild_id: m.GuildID, power: 0}
		if tokens[0] != "/power" && tokens[0] != "/givepower" {
			current_power_data.power = 1
		}
		PowerStructs = append(PowerStructs, current_power_data)
	}

	if tokens[0] == "/power" {
		for _, item := range PowerStructs {
			if item.guild_id == m.GuildID && item.user_id == m.Author.ID {
				s.ChannelMessageSendReply(m.ChannelID, "In this server you currently have: "+strconv.Itoa(item.power)+" power", m.Reference())
			}
		}
	}

	if (tokens[0] == "/givepower" || tokens[0] == "/setpower") && (len(tokens) >= 3) {
		perms, _ := s.State.UserChannelPermissions(m.Author.ID, m.ChannelID)
		guild, _ := s.Guild(m.GuildID)

		if perms&discordgo.PermissionAdministrator == 0 {
			if m.Author.ID != guild.OwnerID {
				s.ChannelMessageSendReply(m.ChannelID, "Sorry you do not have permission to give power to people", m.Reference())
				return
			}
		}

		tokens[1] = strings.Trim(tokens[1], "<@")
		tokens[1] = strings.Trim(tokens[1], ">")

		power_temp := 0

		d, err := s.User(tokens[1])
		if err != nil {
			s.ChannelMessageSendReply(m.ChannelID, "Sorry that is not a valid user", m.Reference())
			return
		}

		found := false

		for i, item := range PowerStructs {
			if item.guild_id == m.GuildID && item.user_id == tokens[1] {
				num_conv, _ := strconv.Atoi(tokens[2])
				if num_conv == 0 {
					s.ChannelMessageSendReply(m.ChannelID, "Sorry that is not a valid number", m.Reference())
					return
				}
				if tokens[0] == "/givepower" {
					PowerStructs[i].power += num_conv
				} else if tokens[0] == "/setpower" {
					PowerStructs[i].power = num_conv
				}
				power_temp = PowerStructs[i].power
				found = true
				break
			}
		}

		if !found {
			num_conv, _ := strconv.Atoi(tokens[2])
			add_user(m.GuildID, tokens[1], num_conv)
			power_temp = num_conv
		}

		s.ChannelMessageSendReply(m.ChannelID, "You gave "+tokens[2]+" power to <@"+d.ID+">\n<@"+d.ID+"> now has "+strconv.Itoa(power_temp)+" power", m.Reference())
	}
}

func add_user(guildid string, userid string, add int) {
	current_power_data := UserPowerData{user_id: userid, guild_id: guildid, power: add}
	PowerStructs = append(PowerStructs, current_power_data)
}
