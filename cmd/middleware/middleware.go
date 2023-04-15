package middleware

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httputil"
	"net/url"
	"path"

	"fiufit.api.gateway/internal/auth"
	"github.com/gin-gonic/gin"
)

const uidKey string = "User-UID"

func ReverseProxy(url *url.URL) gin.HandlerFunc {
	return func(c *gin.Context) {
		proxy := httputil.NewSingleHostReverseProxy(url)
		c.Request.Host = url.Host
		proxy.ServeHTTP(c.Writer, c.Request)
	}
}

func AuthorizeUser(s auth.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Request.Header.Get("Authorization")
		uid, err := s.VerifyToken(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		// Magic value const
		c.Set(uidKey, uid)
	}
}

func AuthorizeAdmin(url *url.URL) gin.HandlerFunc {
	return func(c *gin.Context) {
		UID, ok := getUID(c)
		if !ok {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		adminURL := *url
		adminURL.Path = path.Join(adminURL.Path, "admins", UID)
		resultChannel := make(chan bool)
		go func(rawURL string) {
			response, err := http.Get(rawURL)
			if err != nil || response.StatusCode != http.StatusOK {
				resultChannel <- false
				return
			}
			resultChannel <- true
		}(adminURL.String())

		ok = <-resultChannel
		if !ok {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
	}
}
func AddUIDToRequestURL() gin.HandlerFunc {
	return func(c *gin.Context) {
		UID, ok := getUID(c)
		if !ok {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		// Set the endpoint to request in the users service
		c.Request.URL.Path = path.Join(c.Request.URL.Path, UID)
		// TODO: Remove
		c.Request.Header.Set("Content-Type", "application/json")
	}
}

func getUID(c *gin.Context) (string, bool) {
	anyUID, found := c.Get(uidKey)
	if !found {
		return "", found
	}
	UID, ok := anyUID.(string)
	// Should never fail, dev error
	if !ok {
		return "", ok
	}

	return UID, true
}

// Returns the handler charged with creating an user. It takes the URL
// of the users service and an auth Service as argument.
func CreateUser(s auth.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var signUpData auth.SignUpModel
		// FIX: Doesn't check that all fields are present
		err := c.ShouldBindJSON(&signUpData)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
		req, err := http.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(userDataJSON))
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		req.Header.Set("Content-Type", "application/json")
		c.Request = req
	}
}

//TODO: TEST
func Cors() gin.HandlerFunc{
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		c.Header("Access-Control-Allow-Origin", origin)
		c.Header("Access-Control-Allow-Credentials", "true")
		if c.Request.Method == http.MethodOptions {
			c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type, Content-Length")
			c.Header("Access-Control-Allow-Methods", "POST, GET, PUT, DELETE, OPTIONS")
		}
		c.Status(http.StatusOK)
	}
}
