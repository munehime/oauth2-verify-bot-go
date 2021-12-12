package router

import (
	"net/http"

	"github.com/foolin/goview"
	"github.com/foolin/goview/supports/ginview"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"

	"github.com/munehime/oauth2-verify-bot-go/src/config"
	authorizeController "github.com/munehime/oauth2-verify-bot-go/src/controllers/authorize"
)

type Router struct{}

func MountRouter() *gin.Engine {
	config := config.GetConfig()
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	store := cookie.NewStore([]byte(config.GetString("server.session_key")))
	router.Use(sessions.Sessions("osb", store))
	router.HTMLRender = ginview.New(goview.Config{
		Root: "src/views",
	})

	apiGroup := router.Group("/")
	{
		apiGroup.GET("/", func(ctx *gin.Context) {
			ctx.HTML(http.StatusOK, "index.html", gin.H{})
		})

		authorizeGroup := apiGroup.Group("/authorize")
		{
			authorizeGroup.GET("/osu", authorizeController.HandleOsuAuthorization)
			authorizeGroup.GET("/osu/callback", authorizeController.HandleOsuAuthorizationCallback)
			authorizeGroup.GET("/osu/success", authorizeController.HandleOsuAuthorizationSuccess)
			authorizeGroup.GET("/discord", authorizeController.HandleDiscordAuthorization)
			authorizeGroup.GET("/discord/callback", authorizeController.HandleDiscordAuthorizationCallback)
			authorizeGroup.GET("/discord/success", authorizeController.HandleDiscordAuthorizationSuccess)
		}
	}

	return router
}
