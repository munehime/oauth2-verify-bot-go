package discord

import (
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/munehime/oauth2-verify-bot-go/src/config"
)

var client *discordgo.Session

func Init() {
	config := config.GetConfig()

	var err error
	client, err = discordgo.New("Bot " + config.GetString("discord.token"))
	if err != nil {
		log.Println("Error creating Discord session,", err)
		return
	}

	client.AddHandler(guildMemberAdd)

	client.Identify.Intents = discordgo.IntentsGuildMembers | discordgo.IntentsGuildMessages

	err = client.Open()
	if err != nil {
		log.Println("Error opening connection,", err)
		return
	}

	log.Println("Bot is ready!")
}

func GetClient() *discordgo.Session {
	return client
}

func AddRole(userID string) {
	config := config.GetConfig()

	err := client.GuildMemberRoleAdd(config.GetString("discord.guild"), userID, config.GetString("discord.verifiedRole"))
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Added role to user")
}

func guildMemberAdd(client *discordgo.Session, member *discordgo.GuildMemberAdd) {
	config := config.GetConfig()
	log.Println(config.GetString("discord.welcomeChannel"))
	client.ChannelMessageSend(config.GetString("discord.welcomeChannel"), "Welcome <@"+member.User.ID+">, please go to <"+config.GetString("server.publicUrl")+"> to verify!")
}
