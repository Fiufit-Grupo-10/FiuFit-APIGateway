package main

import (
	"net/http"
	"github.com/gin-gonic/gin"
)

type SignUpData struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthData struct {
	UID string `json:"uid"`
	Token string `json:"token"`
	RefreshToken string `json:"refresh_token"`
}

func main() {
	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {c.JSON(http.StatusOK, "pong!")})
	r.Run() // listen and serve on 0.0.0.0:8080
}
