package main

import (
	"fmt"
	"github.com/FrNecas/GreyaBot/bot"
	"github.com/FrNecas/GreyaBot/config"
)

func main() {
	err := config.ReadConfig()
	if err != nil {
		fmt.Println("Error while creating config,", err)
		return
	}
	bot.RunBot()
}
