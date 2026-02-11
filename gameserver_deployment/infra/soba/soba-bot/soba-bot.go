package main

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

var (
	commands = []*discordgo.ApplicationCommand{
		{
			Name:        "get-server-ip",
			Description: "Get the IP address and port for Noodles dedicated servers",
		},
	}
	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"get-server-ip": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			ip, err := getOutboundIP()
			if err != nil {
				fmt.Println("error getting ip:", err)
				return
			}

			games := parseGamesEnv()
			content := strings.Builder{}
			content.WriteString(fmt.Sprintf("### IP address:\n%s\n### Game Ports:\n", ip))
			for key, value := range games {
				content.WriteString(fmt.Sprintf("- %s: %s\n", key, value))
			}

			err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Flags:   discordgo.MessageFlagsEphemeral,
					Content: content.String(),
				},
			})
			if err != nil {
				if _, err = s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
					Content: "Something went wrong",
				}); err != nil {
					fmt.Println("error creating error follow up...", err)
				}
				return
			}
		},
	}
)

func getOutboundIP() (string, error) {
	url := "https://api.ipify.org?format=text"
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Println("error closing url request:", err)
		}
	}()

	ip, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(ip), nil
}

func parseGamesEnv() map[string]string {
	m := make(map[string]string)
	entries := strings.Split(os.Getenv("GAMES"), ",")

	for _, entry := range entries {
		entry = strings.TrimSpace(entry)
		parts := strings.Split(entry, "=")

		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			m[key] = value
		}
	}
	return m
}

func main() {
	dg, err := discordgo.New("Bot " + os.Getenv("TOKEN"))
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	dg.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
			h(s, i)
		}
	})

	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	registeredCommands := make([]*discordgo.ApplicationCommand, len(commands))
	for i, v := range commands {
		cmd, err := dg.ApplicationCommandCreate(dg.State.User.ID, os.Getenv("GUILD_ID"), v)
		if err != nil {
			log.Panicf("Cannot create '%v' command: %v", v.Name, err)
		}
		registeredCommands[i] = cmd
	}

	fmt.Println("Bot is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	if err = dg.Close(); err != nil {
		fmt.Println("error closing connection,", err)
		return
	}
}
