package main

import (
	"discord_bot/bot"
	"log"
	"os"
)

func main() {
	data, err := os.ReadFile("token.txt")
	if err != nil {
		log.Fatal(err)
	}
	botToken := string(data)
	bot.BotToken = botToken
	bot.Run()
}
