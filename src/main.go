package main

import (
	"github.com/munehime/oauth2-verify-bot-go/src/config"
	"github.com/munehime/oauth2-verify-bot-go/src/database"
	"github.com/munehime/oauth2-verify-bot-go/src/server"
	discordService "github.com/munehime/oauth2-verify-bot-go/src/services/discord"
	oauth2Service "github.com/munehime/oauth2-verify-bot-go/src/services/oauth2"
)

func main() {
	config.Setup("config")
	database.Connect()
	oauth2Service.Init()
	discordService.Init()
	server.Start()
}
