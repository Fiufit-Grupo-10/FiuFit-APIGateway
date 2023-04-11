package gateway

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"path"
	"strconv"

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



// Returns the handler charged with creating an user. It takes the URL
// of the users service and an auth Service as argument.
func createUser(usersService *url.URL, s auth.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var signUpData auth.SignUpModel
		// FIX: Doesn't check that all fields are present
		err := c.ShouldBindJSON(&signUpData)
		if err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		userData, err := s.CreateUser(signUpData)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}

		// Should never fail unless the userData
		// representation becomes an unsupported type
		userDataJSON, err := json.Marshal(userData)
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		
		c.Request.Body = io.NopCloser(bytes.NewReader(userDataJSON))
		c.Request.Header.Set("Content-Length", strconv.Itoa(len(userDataJSON)))
		c.Request.ContentLength = int64(len(userDataJSON))

		proxy :=  httputil.NewSingleHostReverseProxy(usersService)
		proxy.ServeHTTP(c.Writer, c.Request)
		
		c.JSON(http.StatusCreated, userData)
	}
}

// Sets the routes for the users endpoint
func Users(url *url.URL, auth auth.Service) RouterConfig {
	return func(router *gin.Engine) { router.POST("/users", createUser(url, auth)) }
}

func updateProfile(usersService *url.URL, s auth.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		// This is it's own middleware, share with context
		token := c.Request.Header.Get("Authorization")
		uid, err := s.VerifyToken(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		// Set the endpoint to request in the users service
		c.Request.URL.Path = path.Join(c.Request.URL.Path, uid)
		
		proxy :=  httputil.NewSingleHostReverseProxy(usersService)
		proxy.ServeHTTP(c.Writer, c.Request)
	}
}
