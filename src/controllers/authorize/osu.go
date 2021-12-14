package authorize

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/munehime/oauth2-verify-bot-go/src/config"
	"github.com/munehime/oauth2-verify-bot-go/src/database"
	userModel "github.com/munehime/oauth2-verify-bot-go/src/models"
	oauth2Service "github.com/munehime/oauth2-verify-bot-go/src/services/oauth2"
)

type OsuUser struct {
	ID        uint64 `json:"id"`
	Username  string `json:"username"`
	AvatarURL string `json:"avatar_url"`
	Country   string `json:"country_code"`
}

func HandleOsuAuthorization(ctx *gin.Context) {
	ctx.Redirect(http.StatusFound,
		oauth2Service.GetOsuOAuth2Client().AuthCodeURL(""),
	)
}

func HandleOsuAuthorizationCallback(ctx *gin.Context) {
	code := ctx.Query("code")
	if code == "" {
		ctx.Redirect(http.StatusBadRequest, "/")
		return
	}

	token, err := oauth2Service.GetOsuOAuth2Client().Exchange(context.TODO(), code)
	if err != nil {
		log.Errorln(err)
	}

	resty := resty.New()
	resp, err := resty.R().
		SetHeader("Content-Type", "application/json").
		SetAuthToken(token.AccessToken).
		Get("https://osu.ppy.sh/api/v2/me")
	if err != nil {
		log.Errorln(err)
	}

	osuUser := OsuUser{}
	err = json.Unmarshal(resp.Body(), &osuUser)
	if err != nil {
		log.Errorln(err)
	}

	c, cancel := context.WithTimeout(context.TODO(), 10*time.Second)
	defer cancel()

	config := config.GetConfig()
	collection := database.Client().Database(config.GetString("database.name")).Collection("users")
	doc := collection.FindOne(c, bson.M{
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
			CreatedAt: time.Now(),
		}

		result, err := collection.InsertOne(c, user)
		if err != nil {
			log.Errorln(err)
		}

		docID = result.InsertedID.(primitive.ObjectID)
	}

	user.Osu.UserID = strconv.FormatUint(osuUser.ID, 10)
	user.Osu.Username = osuUser.Username
	user.Osu.AvatarURL = osuUser.AvatarURL
	user.Osu.LastVerified = time.Now()

	user.Country = osuUser.Country
	user.LastLogin = time.Now()
	user.UpdatedAt = time.Now()

	if _, err := collection.UpdateOne(c,
		bson.M{"_id": docID},
		bson.M{"$set": user},
	); err != nil {
		log.Errorln(err)
	}

	session := sessions.Default(ctx)
	session.Set("id", user.Osu.UserID)
	session.Set("osu-username", user.Osu.Username)
	session.Save()

	ctx.Redirect(http.StatusFound,
		config.GetString("server.publicUrl")+"/authorize/osu/success",
	)
}

func HandleOsuAuthorizationSuccess(ctx *gin.Context) {
	session := sessions.Default(ctx)
	sessionOsuUsername := session.Get("osu-username")

	ctx.HTML(http.StatusOK, "logged-in.html", gin.H{
		"osuUsername": sessionOsuUsername,
	})
}
