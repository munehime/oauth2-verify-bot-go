package authorize

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/munehime/oauth2-verify-bot-go/src/config"
	"github.com/munehime/oauth2-verify-bot-go/src/database"
	userModel "github.com/munehime/oauth2-verify-bot-go/src/models"
	discordService "github.com/munehime/oauth2-verify-bot-go/src/services/discord"
	oauth2Service "github.com/munehime/oauth2-verify-bot-go/src/services/oauth2"
)

type DiscordUser struct {
	ID            string `json:"id"`
	Username      string `json:"username"`
	Discriminator string `json:"discriminator"`
	Avatar        string `json:"avatar"`
}

func HandleDiscordAuthorization(ctx *gin.Context) {
	session := sessions.Default(ctx)
	sessionID := session.Get("id")

	if sessionID == nil {
		ctx.JSON(401, gin.H{
			"message": "Please login via osu! first",
		})
		return
	}

	ctx.Redirect(http.StatusFound,
		oauth2Service.GetDiscordOAuth2Client().AuthCodeURL(""),
	)
}

func HandleDiscordAuthorizationCallback(ctx *gin.Context) {
	session := sessions.Default(ctx)
	sessionID := session.Get("id")

	if sessionID == nil {
		ctx.JSON(401, gin.H{
			"message": "Please login via osu! first",
		})
		return
	}

	code := ctx.Query("code")
	if code == "" {
		ctx.Redirect(http.StatusBadRequest, "/")
		return
	}

	token, err := oauth2Service.GetDiscordOAuth2Client().Exchange(context.TODO(), code)
	if err != nil {
		log.Errorln(err)
	}

	resty := resty.New()
	resp, err := resty.R().
		SetHeader("Content-Type", "application/json").
		SetAuthToken(token.AccessToken).
		Get("https://discord.com/api/v9/users/@me")
	if err != nil {
		log.Errorln(err)
	}

	discordUser := DiscordUser{}
	err = json.Unmarshal(resp.Body(), &discordUser)
	if err != nil {
		log.Errorln(err)
	}

	c, cancel := context.WithTimeout(context.TODO(), 10*time.Second)
	defer cancel()

	config := config.GetConfig()
	collection := database.Client().Database(config.GetString("database.name")).Collection("users")
	doc := collection.FindOne(c, bson.M{
		"osu.userId": sessionID,
	})

	user := userModel.User{}
	err = doc.Decode(&user)
	if err != nil {
		if err != mongo.ErrNoDocuments {
			log.Errorln(err)
		}

		ctx.JSON(401, gin.H{
			"message": "Please login via osu! first",
		})
	}

	docID := user.ID
	if user.Discord == (userModel.OAuth{}) {
		user.Discord = userModel.OAuth{
			DateAdded: time.Now(),
		}
	}

	user.Discord.UserID = discordUser.ID
	user.Discord.Username = discordUser.Username + "#" + discordUser.Discriminator

	usr, err := discordService.GetClient().User(discordUser.ID)
	if err != nil {
		log.Errorln(err)
	}

	user.Discord.AvatarURL = usr.AvatarURL("")
	user.Discord.LastVerified = time.Now()

	user.LastLogin = time.Now()
	user.UpdatedAt = time.Now()

	if _, err := collection.UpdateOne(c,
		bson.M{"_id": docID},
		bson.M{"$set": user},
	); err != nil {
		log.Errorln(err)
	}

	discordService.AddRole(discordUser.ID)
	discordService.ChangeNickname(discordUser.ID, user.Osu.Username)

	session.Set("discord-username", user.Discord.Username)
	session.Save()

	ctx.Redirect(http.StatusFound,
		config.GetString("server.publicUrl")+"/authorize/discord/success",
	)
}

func HandleDiscordAuthorizationSuccess(ctx *gin.Context) {
	session := sessions.Default(ctx)
	sessionOsuUsername := session.Get("osu-username")
	sessionDiscordUsername := session.Get("discord-username")

	ctx.HTML(http.StatusOK, "authorized.html", gin.H{
		"osuUsername":     sessionOsuUsername,
		"discordUsername": sessionDiscordUsername,
	})
}
