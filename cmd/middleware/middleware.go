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
	return gin.WrapH(httputil.NewSingleHostReverseProxy(url))
}

func Authorize(s auth.Service) gin.HandlerFunc {
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

func AddUIDToRequestURL() gin.HandlerFunc {
	return func(c *gin.Context) {
		anyUID, found := c.Get(uidKey)
		if !found {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		UID, ok := anyUID.(string)
		// Should never fail, dev error
		if !ok {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		// Set the endpoint to request in the users service
		c.Request.URL.Path = path.Join(c.Request.URL.Path, UID)
	}
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
		req, err := http.NewRequest(http.MethodPost, "/users",bytes.NewBuffer(userDataJSON))
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		req.Header.Set("Content-Type", "application/json")
		c.Request = req
	}
}
