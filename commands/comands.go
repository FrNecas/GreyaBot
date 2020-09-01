package commands

import (
	"fmt"
	"github.com/FrNecas/GreyaBot/config"
	"github.com/bwmarrin/discordgo"
)

func Execute(s *discordgo.Session, m *discordgo.Message) {
	UnknownCommand(s, m)
}

func UnknownCommand(s *discordgo.Session, m *discordgo.Message) {
	_, err := s.ChannelMessageSend(m.ChannelID, config.Config.UnknownCommandMessage)
	if err != nil {
		fmt.Println("Error while handling unknown command")
	}
}
