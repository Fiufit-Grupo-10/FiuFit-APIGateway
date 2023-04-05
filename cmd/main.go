package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type UserModel struct {
	UID      string `json:"uid"`
	Email    string `json:"email"`
	Username string `json:"username"`
}

func main() {
	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) { c.JSON(http.StatusOK, "pong!") })
	r.Run() // listen and serve on 0.0.0.0:8080
}
