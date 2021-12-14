package osu

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/munehime/oauth2-verify-bot-go/src/config"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2/clientcredentials"
)

var (
	clientConfig *clientcredentials.Config
	osuClient    *http.Client
)

type OsuUser struct {
	ID        uint64 `json:"id"`
	Username  string `json:"username"`
	AvatarURL string `json:"avatar_url"`
	Country   string `json:"country_code"`
	Discord   string `json:"discord"`
}

func Init() {
	config := config.GetConfig()

	clientConfig = &clientcredentials.Config{
		ClientID:     config.GetString("osu.v2.clientId"),
		ClientSecret: config.GetString("osu.v2.clientSecret"),
		TokenURL:     "https://osu.ppy.sh/oauth/token",
		Scopes:       []string{"public"},
	}

	osuClient = clientConfig.Client(context.TODO())
}

func GetOsuProfile(userID string) (OsuUser, error) {
	request, _ := http.NewRequest("GET", "https://osu.ppy.sh/api/v2/users/"+userID, nil)
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Content-Type", "application/json")
	response, err := osuClient.Do(request)
	if err != nil {
		log.Errorln(err)
		return OsuUser{}, err
	}

	user := OsuUser{}
	err = json.NewDecoder(response.Body).Decode(&user)
	if err != nil {
		log.Errorln(err)
		return OsuUser{}, err
	}

	return user, nil
}
