package commands

import (
	"fmt"
	"github.com/FrNecas/GreyaBot/config"
	"github.com/FrNecas/GreyaBot/db"
	"github.com/FrNecas/GreyaBot/message"
	"github.com/bwmarrin/discordgo"
	"strconv"
	"strings"
)

func updateChannelSettings(s *discordgo.Session, channelID string, limit int) {
	channel, err := s.Channel(channelID)
	if err != nil {
		fmt.Println("Couldn't retrieve voice channel")
		return
	}
	settings := discordgo.ChannelEdit{
		Name:                 channel.Name,
		Topic:                channel.Topic,
		NSFW:                 channel.NSFW,
		Position:             channel.Position,
		Bitrate:              channel.Bitrate,
		UserLimit:            limit,
		PermissionOverwrites: channel.PermissionOverwrites,
		ParentID:             channel.ParentID,
		RateLimitPerUser:     channel.RateLimitPerUser,
	}
	_, err = s.ChannelEditComplex(channelID, &settings)
	if err != nil {
		fmt.Println("Error changing user limit in channel,", err)
	}
}

func setVoiceLimit(s *discordgo.Session, m *discordgo.Message, limit string) {
	// Get the owned channel
	database, err := db.Connect()
	if err != nil {
		fmt.Println("Couldn't connect to the database when updating voice channel,", err)
		return
	}
	rows, err := database.Query(`SELECT channel_id FROM voice_channels WHERE owner_id = $1`, m.Author.ID)
	if err != nil {
		fmt.Println("Error when querying data when updating channel,", err)
	}
	var channels []string
	for rows.Next() {
		var channelID string
		err := rows.Scan(&channelID)
		if err != nil {
			fmt.Println("Error when scanning data when updating channel,", err)
		}
		channels = append(channels, channelID)
	}
	if len(channels) == 0 {
		s.ChannelMessageSend(m.ChannelID, config.Config.NoChannelOwnedMessage)
		return
	}

	// Convert the given limit to a number
	intLimit, err := strconv.Atoi(limit)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, config.Config.LimitMustBeNumberMessage)
	}
	updateChannelSettings(s, channels[0], intLimit)
}

func Execute(s *discordgo.Session, m *discordgo.Message) {
	msg := strings.TrimPrefix(m.Content, config.Config.BotPrefix)
	withoutDiacritics := strings.ToLower(message.RemoveDiacritics(msg))
	split := strings.Split(withoutDiacritics, " ")
	if len(split) >= 2 && split[0] == "voice" {
		// The voice commands are always in the form !voice <command> <possibly something else>
		if split[1] == "setlimit" && len(split) >= 3 {
			setVoiceLimit(s, m, split[2])
		} else {
			UnknownCommand(s, m)
		}
		return
	}
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
