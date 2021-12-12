package server

import (
	"github.com/munehime/oauth2-verify-bot-go/src/config"
	"github.com/munehime/oauth2-verify-bot-go/src/router"
)

func Start() {
	config := config.GetConfig()
	router := router.MountRouter()
	router.Static("/static", "./static")
	router.RunTLS(":"+config.GetString("server.port"),
		"/home/mune/cert/server.crt",
		"/home/mune/cert/server.key",
	)
}
