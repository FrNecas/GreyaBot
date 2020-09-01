package commands

import (
	"fmt"
	"github.com/FrNecas/GreyaBot/config"
	"github.com/FrNecas/GreyaBot/message"
	"github.com/bwmarrin/discordgo"
	"strings"
)

func Execute(s *discordgo.Session, m *discordgo.Message) {
	msg := strings.TrimPrefix(m.Content, config.Config.BotPrefix)
	withoutDiacritics := message.RemoveDiacritics(msg)
	switch withoutDiacritics {
	// More special and complex commands can be added here
	default:
		if !HandleSimpleCommands(s, m, withoutDiacritics) {
			UnknownCommand(s, m)
		}
	}
}

func HandleSimpleCommands(s *discordgo.Session, m *discordgo.Message, withoutDiacritics string) bool {
	for _, cmd := range config.Config.PredefinedCommands {
		if cmd.Command == withoutDiacritics {
			_, err := s.ChannelMessageSend(m.ChannelID, cmd.Reaction)
			if err != nil {
				fmt.Println("Error while sending reaction to a simple command")
			}
			return true
		}
	}
	return false
}

func UnknownCommand(s *discordgo.Session, m *discordgo.Message) {
	_, err := s.ChannelMessageSend(m.ChannelID, config.Config.UnknownCommandMessage)
	if err != nil {
		fmt.Println("Error while handling unknown command")
	}
}
