package gateway

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"

	"fiufit.api.gateway/internal/auth"
	"github.com/gin-gonic/gin"
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

func reverseProxy(url *url.URL) gin.HandlerFunc {
	return func(c *gin.Context) {
		proxy := httputil.NewSingleHostReverseProxy(url)
		proxy.ServeHTTP(c.Writer, c.Request)
	}
}

func createUser(usersService *url.URL, s auth.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var signUpData auth.SignUpModel
		err := c.ShouldBindJSON(&signUpData)
		if err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		//TODO: This error should send the context to the client
		userData, err := s.CreateUser(signUpData)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}

		userDataJSON, err := json.Marshal(userData)
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		resultChannel := make(chan error)
		go func(url string, body io.Reader) {
			response, err := http.Post(url, "application/json", body)
			response.Body.Close()
			resultChannel <- err
		}(usersService.String(), bytes.NewReader(userDataJSON))

		err = <-resultChannel
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		// Handle this better
		c.JSON(http.StatusCreated, userData)
	}
}

func Users(url *url.URL, auth auth.Service) RouterConfig {
	return func(router *gin.Engine) { router.POST("/users", createUser(url, auth)) }
}
