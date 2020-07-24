package main

import (
	"fmt"
	"github.com/FrNecas/GreyaBot/bot"
	"github.com/FrNecas/GreyaBot/config"
	"github.com/FrNecas/GreyaBot/twitch"
)

func main() {
	err := config.ReadConfig()
	if err != nil {
		fmt.Println("Error while creating config,", err)
		return
	}
	msgChannel := make(chan string)
	twitch.StartServerGoroutine(msgChannel)
	bot.RunBot(msgChannel)
}
