package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"path"
	"strings"

	"fiufit.api.gateway/internal/auth"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

const uidKey string = "User-UID"
const authorizedKey string = "Authorized"
const allowedHeaders string = "Authorization, Content-Type, Content-Length"
const allowedMethods string = "POST, GET, PUT, DELETE, OPTIONS, PATCH"

type BlockModel struct {
	UID     string `json:"uid"`
	Blocked bool   `json:"blocked"`
}

func ChangeBlockStatusFirebase(s auth.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Conseguir uid y blocked,
		// Tratar de bloquearlos
		var users []BlockModel
		err := c.BindJSON(&users)
		if err != nil {
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

		buf, err := json.Marshal(users)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, err.Error())
			return
		}

		c.Request.Body = io.NopCloser(bytes.NewBuffer(buf))
	}
}

func SetQuery(key, value string) gin.HandlerFunc {
	return func(c *gin.Context) {
		query := c.Request.URL.Query()
		query.Add(key, value)
		c.Request.URL.RawQuery = query.Encode()
	}
}

func ReverseProxy(url *url.URL) gin.HandlerFunc {
	return func(c *gin.Context) {
		proxy := httputil.NewSingleHostReverseProxy(url)
		clientIp := c.ClientIP()
		proxy.ErrorHandler = getErrorHandler(clientIp)
		c.Request.Host = url.Host
		proxy.ServeHTTP(c.Writer, c.Request)
	}
}

func getErrorHandler(clientIP string) func(http.ResponseWriter, *http.Request, error) {
	return func(rw http.ResponseWriter, r *http.Request, e error) {
		log.WithFields(log.Fields{"uri": r.RequestURI, "client_ip": clientIP, "error": e.Error()}).Info("Reverse proxy failed")
		rw.WriteHeader(http.StatusBadGateway)
	}
}

func AuthorizeUser(s auth.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Request.Header.Get("Authorization")
		uid, err := s.VerifyToken(token)
		logContext := log.Fields{
			"method":    c.Request.Method,
			"client_ip": c.ClientIP(),
			"uri":       c.Request.RequestURI,
		}

		if err != nil {
			logContext["authorized"] = true
			logContext["error"] = err.Error()
			log.WithFields(logContext).Info("Firebase Authorization failed")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		logContext["authorized"] = true
		log.WithFields(logContext).Info("Firebase Authorization done")
		c.Set(uidKey, uid)
	}
}

func AuthorizeAdmin(url *url.URL) gin.HandlerFunc {
	return func(c *gin.Context) {

		UID, ok := getUID(c)
		if !ok {
			log.WithFields(log.Fields{"error": "UID not set in context"}).Error("Admin authentication failed")
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
			log.WithFields(log.Fields{"error": "Not an admin"}).Info("Admin authentication failed")
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

// Returns the handler charged with creating an user. It takes the URL
// of the users service and an auth Service as argument.
// TODO: Delete user
func CreateUser(s auth.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var signUpData auth.SignUpModel
		// FIX: Doesn't check that all fields are present
		err := c.ShouldBindJSON(&signUpData)
		if err != nil {
			log.WithFields(log.Fields{"error": err.Error()}).Info("couldn't bind to json sign up form")
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var userData auth.UserModel
		if signUpData.Federated {
			userData, err = s.GetUser(signUpData.UID)
		} else {
			userData, err = s.CreateUser(signUpData)
		}

		log.WithFields(log.Fields{"user": userData}).Info("Creating user in firebase")
		if err != nil {
			log.WithFields(log.Fields{"user": userData}).Info("Failed to create user in firebase")
			c.AbortWithStatusJSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		// Should never fail unless the userData
		// representation becomes an unsupported type
		userDataJSON, err := json.Marshal(userData)
		if err != nil {
			// Delete user from firebase
			log.WithFields(log.Fields{"data": userData, "error": err.Error()}).Error("couldn't marshall data to json ")
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		log.WithFields(log.Fields{"user": string(userDataJSON)}).Info("initialized user in users service")
		req, err := http.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(userDataJSON))
		if err != nil {
			// delete user from firebase
			log.WithFields(log.Fields{"error": err.Error(), "user": string(userDataJSON)}).Error("couldn't reach users service")
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
			log.WithFields(log.Fields{"error": err.Error()}).Info("couldn't bind to json sign up form")
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		userData, err := s.CreateUser(signUpData)
		log.WithFields(log.Fields{"admin": userData}).Info("Creating admin in firebase")
		if err != nil {
			log.WithFields(log.Fields{"user": userData}).Info("Failed to create admin in firebase")
			c.AbortWithStatusJSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}

		// Should never fail unless the userData
		// representation becomes an unsupported type
		userDataJSON, err := json.Marshal(userData)
		if err != nil {
			log.WithFields(log.Fields{"data": userData, "error": err.Error()}).Error("couldn't marshall data to json ")
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		log.WithFields(log.Fields{"admin": string(userDataJSON)}).Info("initialized admin in users service")
		req, err := http.NewRequest(http.MethodPost, "/admins", bytes.NewBuffer(userDataJSON))
		if err != nil {
			log.WithFields(log.Fields{"error": err.Error(), "user": string(userDataJSON)}).Error("couldn't reach users service")
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

func Logger() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Request method
		requestMethod := ctx.Request.Method

		// Request route
		requestURI := ctx.Request.RequestURI

		// status code
		statusCode := ctx.Writer.Status()

		// Request IP
		clientIP := ctx.ClientIP()

		log.WithFields(log.Fields{
			"method":    requestMethod,
			"uri":       requestURI,
			"status":    statusCode,
			"client_ip": clientIP,
		}).Info("HTTP Request")

	}
}
