package discord

import (
	"context"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/munehime/oauth2-verify-bot-go/src/config"
	"github.com/munehime/oauth2-verify-bot-go/src/database"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	userModel "github.com/munehime/oauth2-verify-bot-go/src/models"
	osuService "github.com/munehime/oauth2-verify-bot-go/src/services/osu"
)

var client *discordgo.Session

const REGEX_URL string = `(?m)^(?:https?://)?(?:[^/.\s]+\.)*osu\.ppy\.sh\/users(?:/[^/\s]+)*/?$`
const REGEX_MENTION = `(?m)^<@!?(\d+)>$`

func Init() {
	config := config.GetConfig()

	var err error
	client, err = discordgo.New("Bot " + config.GetString("discord.token"))
	if err != nil {
		log.Fatalln("Error creating Discord session,", err)
		return
	}

	client.AddHandler(guildMemberAdd)
	client.AddHandler(messageCreate)

	client.Identify.Intents = discordgo.IntentsGuildMembers | discordgo.IntentsGuildMessages

	err = client.Open()
	if err != nil {
		log.Fatalln("Error opening connection,", err)
		return
	}

	log.Infoln("Bot is ready!")
}

func GetClient() *discordgo.Session {
	return client
}

func AddRole(userID string) {
	config := config.GetConfig()

	err := client.GuildMemberRoleAdd(config.GetString("discord.guild"), userID, config.GetString("discord.verifiedRole"))
	if err != nil {
		log.Errorln(err)
	}

	log.Infoln("Added role to user")
}

func ChangeNickname(userID string, nickname string) {
	config := config.GetConfig()

	err := client.GuildMemberNickname(config.GetString("discord.guild"), userID, nickname)
	if err != nil {
		log.Errorln(err)
	}

	log.Infoln("Updated user's nickname")
}

func guildMemberAdd(client *discordgo.Session, member *discordgo.GuildMemberAdd) {
	config := config.GetConfig()

	_, err := client.ChannelMessageSend(config.GetString("discord.welcomeChannel"), "Welcome <@"+member.User.ID+">, please go to <"+config.GetString("server.publicUrl")+"> or send you could send your osu! profile link here to verify!\n**Enjoy your stay!**")
	if err != nil {
		log.Errorln(err)
	}
}

func messageCreate(client *discordgo.Session, message *discordgo.MessageCreate) {
	if message.Author.Bot {
		return
	}

	config := config.GetConfig()
	guildID := config.GetString("discord.guild")

	if message.GuildID != guildID {
		return
	}

	if match, _ := regexp.MatchString(REGEX_URL, message.Content); match {
		member, err := client.GuildMember(guildID, message.Author.ID)
		if err != nil {
			log.Errorln(err)
			return
		}

		verifiedRoleID := config.GetString("discord.verifiedRole")
		isVerified := false
		for _, role := range member.Roles {
			if role == verifiedRoleID {
				isVerified = true
				break
			}
		}

		if !isVerified {
			user := path.Base(message.Content)
			verifyUser(message, user)
		}
	}

	prefix := config.GetString("discord.prefix")

	if !strings.HasPrefix(message.Content, prefix) {
		return
	}

	args := regexp.MustCompile(`(?m) +`).Split(strings.TrimSpace(strings.TrimPrefix(message.Content, prefix)), -1)

	if args[0] == "verify" {
		member, err := client.GuildMember(guildID, message.Author.ID)
		if err != nil {
			log.Errorln(err)
			return
		}

		moderatorRoles := config.GetStringSlice("discord.moderatorRoles")
		isModerator := false

		for _, role := range member.Roles {
			for _, moderatorRole := range moderatorRoles {
				if role == moderatorRole {
					isModerator = true
					break
				}
			}
		}

		if match, _ := regexp.MatchString(REGEX_MENTION, args[1]); match && isModerator {
			user := path.Base(args[2])
			memberID := strings.TrimRight(strings.TrimLeft(strings.TrimLeft(args[1], "<@"), "!"), ">")

			verifyUserManually(message, memberID, user)
		}
	}
}

func verifyUser(message *discordgo.MessageCreate, u string) {
	config := config.GetConfig()

	osuUser, err := osuService.GetOsuProfile(u)
	if err != nil {
		client.ChannelMessageSend(message.ChannelID, "Error while verifying the user")
		return
	}

	usr, err := client.User(message.Author.ID)
	if err != nil {
		log.Errorln(err)
	}

	if osuUser.Discord == "" {
		client.ChannelMessageSend(message.ChannelID, "Please put your Discord on your osu! profile")
		return
	}

	userTag := message.Author.Username + "#" + message.Author.Discriminator
	if userTag != osuUser.Discord {
		client.ChannelMessageSend(message.ChannelID, "Your Discord tag does not match")
		return
	}

	ctx, cancel := context.WithTimeout(context.TODO(), 10*time.Second)
	defer cancel()

	collection := database.Client().Database(config.GetString("database.name")).Collection("users")
	doc := collection.FindOne(ctx, bson.M{
		"osu.userId": strconv.FormatUint(osuUser.ID, 10),
	})

	user := userModel.User{}
	err = doc.Decode(&user)
	if err != nil {
		if err != mongo.ErrNoDocuments {
			log.Errorln(err)
		}
	}

	docID := user.ID
	if docID == primitive.NilObjectID {
		user = userModel.User{
			Osu: userModel.OAuth{
				DateAdded: time.Now(),
			},
			Discord: userModel.OAuth{
				DateAdded: time.Now(),
			},
			CreatedAt: time.Now(),
		}

		result, err := collection.InsertOne(ctx, user)
		if err != nil {
			log.Errorln(err)
		}

		docID = result.InsertedID.(primitive.ObjectID)
	}

	user.Osu.UserID = strconv.FormatUint(osuUser.ID, 10)
	user.Osu.Username = osuUser.Username
	user.Osu.AvatarURL = osuUser.AvatarURL
	user.Osu.LastVerified = time.Now()

	user.Discord.UserID = message.Author.ID
	user.Discord.Username = userTag

	user.Discord.AvatarURL = usr.AvatarURL("")
	user.Discord.LastVerified = time.Now()

	user.Country = osuUser.Country
	user.LastLogin = time.Now()
	user.UpdatedAt = time.Now()

	AddRole(message.Author.ID)
	ChangeNickname(message.Author.ID, user.Osu.Username)

	if _, err := collection.UpdateOne(ctx,
		bson.M{"_id": docID},
		bson.M{"$set": user},
	); err != nil {
		log.Errorln(err)
	}

	client.ChannelMessageSend(message.ChannelID, "You're verified now <@"+message.Author.ID+">!")
}

func verifyUserManually(message *discordgo.MessageCreate, memberID string, u string) {
	config := config.GetConfig()

	osuUser, err := osuService.GetOsuProfile(u)
	if err != nil {
		client.ChannelMessageSend(message.ChannelID, "Error while verifying the user")
		return
	}

	usr, err := client.User(memberID)
	if err != nil {
		log.Errorln(err)
	}

	if osuUser.Discord == "" {
		client.ChannelMessageSend(message.ChannelID, "Please tell the user to put his/her Discord on his/her osu! profile")
		return
	}

	userTag := usr.Username + "#" + usr.Discriminator
	if userTag != osuUser.Discord {
		client.ChannelMessageSend(message.ChannelID, "His/Her Discord tag does not match")
		return
	}

	ctx, cancel := context.WithTimeout(context.TODO(), 10*time.Second)
	defer cancel()

	collection := database.Client().Database(config.GetString("database.name")).Collection("users")
	doc := collection.FindOne(ctx, bson.M{
		"osu.userId": strconv.FormatUint(osuUser.ID, 10),
	})

	user := userModel.User{}
	err = doc.Decode(&user)
	if err != nil {
		if err != mongo.ErrNoDocuments {
			log.Errorln(err)
		}
	}

	docID := user.ID
	if docID == primitive.NilObjectID {
		user = userModel.User{
			Osu: userModel.OAuth{
				DateAdded: time.Now(),
			},
			Discord: userModel.OAuth{
				DateAdded: time.Now(),
			},
			CreatedAt: time.Now(),
		}

		result, err := collection.InsertOne(ctx, user)
		if err != nil {
			log.Errorln(err)
		}

		docID = result.InsertedID.(primitive.ObjectID)
	}

	user.Osu.UserID = strconv.FormatUint(osuUser.ID, 10)
	user.Osu.Username = osuUser.Username
	user.Osu.AvatarURL = osuUser.AvatarURL
	user.Osu.LastVerified = time.Now()

	user.Discord.UserID = usr.ID
	user.Discord.Username = userTag

	user.Discord.AvatarURL = usr.AvatarURL("")
	user.Discord.LastVerified = time.Now()

	user.Country = osuUser.Country
	user.LastLogin = time.Now()
	user.UpdatedAt = time.Now()

	AddRole(usr.ID)
	ChangeNickname(usr.ID, user.Osu.Username)

	if _, err := collection.UpdateOne(ctx,
		bson.M{"_id": docID},
		bson.M{"$set": user},
	); err != nil {
		log.Errorln(err)
	}

	client.ChannelMessageSend(message.ChannelID, "Successfully verified <@"+usr.ID+">!")
}
