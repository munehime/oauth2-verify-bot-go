package oauth2

import (
	"golang.org/x/oauth2"

	"github.com/munehime/oauth2-verify-bot-go/src/config"
)

type AuthorizationResponse struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    string `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
}

var (
	osu     oauth2.Config
	discord oauth2.Config
)

func Init() {
	config := config.GetConfig()

	osu = oauth2.Config{
		ClientID:     config.GetString("osu.v2.clientId"),
		ClientSecret: config.GetString("osu.v2.clientSecret"),
		Scopes:       []string{"identify"},
		RedirectURL:  config.GetString("server.publicUrl") + "/authorize/osu/callback",
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://osu.ppy.sh/oauth/authorize",
			TokenURL: "https://osu.ppy.sh/oauth/token",
		},
	}

	discord = oauth2.Config{
		ClientID:     config.GetString("discord.clientId"),
		ClientSecret: config.GetString("discord.clientSecret"),
		Scopes:       []string{"identify"},
		RedirectURL:  config.GetString("server.publicUrl") + "/authorize/discord/callback",
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://discord.com/api/v9/oauth2/authorize",
			TokenURL: "https://discord.com/api/v9/oauth2/token",
		},
	}
}

func GetOsuOAuth2Client() *oauth2.Config {
	return &osu
}

func GetDiscordOAuth2Client() *oauth2.Config {
	return &discord
}
