package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"path"
	"strings"

	"fiufit.api.gateway/internal/auth"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

const uidKey string = "User-UID"
const authorizedKey string = "Authorized"
const allowedHeaders string = "Authorization, Content-Type, Content-Length"
const allowedMethods string = "POST, GET, PUT, DELETE, OPTIONS, PATCH"

type BlockModel struct {
	UID     string `json:"uid" binding:"required"`
	Blocked bool   `json:"blocked" binding:"required"`
}

func ExecuteIf(guard func(*gin.Context) bool, a, b gin.HandlerFunc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if guard(ctx) {
			a(ctx)
			return
		}
		b(ctx)
	}
}

func ChangeBlockStatusFirebase(s auth.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Conseguir uid y blocked,
		// Tratar de bloquearlos
		var users []BlockModel
		err := c.ShouldBindBodyWith(&users, binding.JSON)
		if err != nil {
			log.Println(err.Error())
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		for _, user := range users {
			_, err := s.GetUser(user.UID)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusNotFound, err.Error())
				return
			}
		}
		
		for _, user := range users {
			// Shouldn't fail
			err = s.SetBlockStatus(user.UID, user.Blocked)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, err.Error())
				return
			}
		}
	}
}

func SetQuery(key, value string) gin.HandlerFunc {
	return func(c *gin.Context) {
		query := c.Request.URL.Query()
		query.Add(key, value)
		c.Request.URL.RawQuery = query.Encode()
	}
}

func IsAuthorized(ctx *gin.Context) bool {
	authorized, found := getAuthorized(ctx)
	// This shouldn't fail, unless it was called incorrectly
	if !found {
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return false
	}
	return authorized
}

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
			c.Set(authorizedKey, false)
			return
		}
		// Magic value const
		c.Set(uidKey, uid)
		c.Set(authorizedKey, true)
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

func getAuthorized(c *gin.Context) (bool, bool) {
	anyAuthorized, found := c.Get(authorizedKey)
	if !found {
		return false, found
	}
	authorized, ok := anyAuthorized.(bool)
	// Should never fail, dev error
	if !ok {
		return false, ok
	}

	return authorized, true
}

// Returns the handler charged with creating an user. It takes the URL
// of the users service and an auth Service as argument.
// TODO: Delete user
func CreateUser(s auth.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var signUpData auth.SignUpModel
		// FIX: Doesn't check that all fields are present
		err := c.ShouldBindJSON(&signUpData)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var userData auth.UserModel
		if signUpData.Federated {
			userData, err = s.GetUser(signUpData.UID)
		} else {
			userData, err = s.CreateUser(signUpData)
		}

		if err != nil {
			c.AbortWithStatusJSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}

		// Should never fail unless the userData
		// representation becomes an unsupported type
		userDataJSON, err := json.Marshal(userData)
		if err != nil {
			// Delete user from firebase
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		req, err := http.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(userDataJSON))
		if err != nil {
			// delete user from firebase
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		c.Request = req

		// C.next()
		// delete user from firebase
	}
}

//TODO: Duplicate code with CreateUser
func CreateAdmin(s auth.Service) gin.HandlerFunc {
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
		req, err := http.NewRequest(http.MethodPost, "/admins", bytes.NewBuffer(userDataJSON))
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		c.Request = req
	}
}

func Cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		c.Header("Access-Control-Allow-Origin", origin)
		c.Header("Access-Control-Allow-Credentials", "true")
		if c.Request.Method == http.MethodOptions {
			c.Header("Access-Control-Allow-Headers", allowedHeaders)
			c.Header("Access-Control-Allow-Methods", allowedMethods)
			c.AbortWithStatus(http.StatusOK)
			return
		}
	}
}

func RemovePathFromRequestURL(path string) gin.HandlerFunc {
	return func(c *gin.Context) {
		url := *c.Request.URL
		splitResult := strings.SplitAfter(url.Path, path)
		if len(splitResult) < 2 {
			return
		}

		if splitResult[1] == "" {
			return
		}

		newURLPath := splitResult[1]
		c.Request.URL.Path = newURLPath
	}
}

func AbortIfNotAuthorized(ctx *gin.Context) {
	authorized, found := getAuthorized(ctx)
	//should never fail dev error
	if !found {
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	if !authorized {
		ctx.AbortWithStatus(http.StatusUnauthorized)
	}
}
