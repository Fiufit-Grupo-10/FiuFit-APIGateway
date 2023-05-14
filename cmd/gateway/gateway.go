package gateway

import (
	"fiufit.api.gateway/cmd/middleware"
	"fiufit.api.gateway/internal/auth"
	"github.com/gin-gonic/gin"
	"net/http"
	"net/url"
)

type RouterConfig func(*gin.Engine)

type Gateway struct {
	router *gin.Engine
}

func (g *Gateway) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	g.router.ServeHTTP(w, r)
}

func (g *Gateway) Run(addr ...string) {
	g.router.Run(addr...)
}

func New(configs ...RouterConfig) *Gateway {
	router := gin.Default()
	router.Use(middleware.Cors())
	for _, option := range configs {
		option(router)
	}
	return &Gateway{router}
}

// Sets the routes for the users endpoint
func Users(url *url.URL, s auth.Service) RouterConfig {
	return func(router *gin.Engine) {

		router.POST("/users", middleware.CreateUser(s), middleware.ReverseProxy(&*url))

		router.GET("/users", middleware.AuthorizeUser(s),
			middleware.ExecuteIf(middleware.IsAuthorized,
				middleware.AddUIDToRequestURL(),
				middleware.SetQuery("admin", "false")),
			middleware.ReverseProxy(&*url))

		router.PUT("/users/:user_id", middleware.AuthorizeUser(s),
			// Verify that it's the same user
			middleware.AbortIfNotAuthorized,
			middleware.ReverseProxy(&*url))

		router.GET("/trainingtypes", middleware.ReverseProxy(&*url))
	}
}

func Admin(url *url.URL, s auth.Service) RouterConfig {
	return func(router *gin.Engine) {
		router.POST("/admins",
			middleware.AuthorizeUser(s),
			middleware.AbortIfNotAuthorized,
			middleware.AuthorizeAdmin(&*url),
			middleware.CreateAdmin(s),
			middleware.ReverseProxy(&*url))

		router.GET("/admins/users",
			middleware.AuthorizeUser(s),
			middleware.AbortIfNotAuthorized,
			middleware.AuthorizeAdmin(&*url),
			middleware.RemovePathFromRequestURL("/admins"),
			middleware.SetQuery("admin", "true"),
			middleware.ReverseProxy(&*url))
	}
}

func Trainers(url *url.URL, s auth.Service) RouterConfig {
	return func(router *gin.Engine) {
		router.POST("/plans",
			middleware.AuthorizeUser(s),
			// Verify that the user is indeed a trainer
			middleware.AbortIfNotAuthorized,
			middleware.ReverseProxy(&*url))

		router.GET("/plans",
			middleware.AuthorizeUser(s),
			middleware.AbortIfNotAuthorized,
			middleware.ReverseProxy(&*url))

		router.PUT("/plans/:plan_id",
			middleware.AuthorizeUser(s),
			// Verify that the user is indeed a trainer
			middleware.AbortIfNotAuthorized,
			middleware.ReverseProxy(&*url))

		router.GET("/plans/:trainer_id",
			middleware.AuthorizeUser(s),
			// Verify that the user is indeed a trainer, and that it's the same
			middleware.AbortIfNotAuthorized,
			middleware.ReverseProxy(&*url))
	}
}

func Reviews(url *url.URL, s auth.Service) RouterConfig {
	return func(router *gin.Engine) {
		router.POST("/reviews",
			middleware.AuthorizeUser(s),
			middleware.AbortIfNotAuthorized,
			middleware.ReverseProxy(&*url))

		router.GET("/reviews/:plan_id",
			middleware.AuthorizeUser(s),
			// Verify that the user is indeed a trainer, and that it's the same
			middleware.AbortIfNotAuthorized,
			middleware.ReverseProxy(&*url))
	}
}
