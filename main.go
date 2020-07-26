package main

import (
	"fmt"
	"github.com/FrNecas/GreyaBot/bot"
	"github.com/FrNecas/GreyaBot/config"
	"github.com/FrNecas/GreyaBot/twitch"
	"github.com/bwmarrin/discordgo"
)

func main() {
	err := config.ReadConfig()
	if err != nil {
		fmt.Println("Error while creating config,", err)
		return
	}
	msgChannel := make(chan *discordgo.MessageSend)
	twitch.StartServerGoroutine(msgChannel)
	bot.RunBot(msgChannel)
}
