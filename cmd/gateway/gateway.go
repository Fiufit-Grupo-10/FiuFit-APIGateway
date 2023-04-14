package gateway

import (
	"fiufit.api.gateway/cmd/middleware"
	"fiufit.api.gateway/internal/auth"
	"github.com/gin-gonic/gin"
	"net/http"
	"net/url"
)

type AuthService interface {
	CreateUser(data auth.SignUpModel) (auth.AuthorizedModel, error)
}

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
	for _, option := range configs {
		option(router)
	}
	return &Gateway{router}
}

// Sets the routes for the users endpoint
func Users(url *url.URL, auth auth.Service) RouterConfig {
	return func(router *gin.Engine) {
		proxy := middleware.ReverseProxy(url)
		router.POST("/users", middleware.CreateUser(auth), proxy)
		router.GET("/users", middleware.AuthorizeUser(auth), middleware.AddUIDToRequestURL(), proxy)
		router.PUT("/users", middleware.AuthorizeUser(auth), middleware.AddUIDToRequestURL(), proxy)
	}
}

func Admin(url *url.URL, auth auth.Service) RouterConfig {
	return func(router *gin.Engine) {
		router.POST("/admins", middleware.AuthorizeUser(auth), middleware.AuthorizeAdmin(url), middleware.CreateUser(auth), middleware.ReverseProxy(url))
		router.OPTIONS("/admins", func(c *gin.Context) {
			c.Header("Allow", "OPTIONS, POST")
		})
	}
}
