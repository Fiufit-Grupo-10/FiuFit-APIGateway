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
	// router.NoRoute(func(c *gin.Context) {
	// 	c.JSON(http.StatusNotFound, gin.H{"code": "PAGE_NOT_FOUND", "message": "404 not found"})
	// })
	router.Use(middleware.Cors())
	for _, option := range configs {
		option(router)
	}
	return &Gateway{router}
}

// Sets the routes for the users endpoint
func Users(url *url.URL, s auth.Service) RouterConfig {
	return func(router *gin.Engine) {
		router.POST("/users", CreateUser(url, s))
		router.GET("/users", GetAuthorizedUserProfile(url, s))
		router.PUT("/users", UpdateUserProfile(url, s))
	}
}

func Admin(url *url.URL, s auth.Service) RouterConfig {
	return func(router *gin.Engine) {
		router.POST("/admins", CreateAdmin(url, s))
		router.GET("/admins/users", GetAllUserProfiles(url, s))
	}
}
