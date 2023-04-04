package main

import (
	"log"
	"net/url"
	"os"

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
	rawURL, found := os.LookupEnv("SERVICE_URL")
	if !found {
		log.Fatal("Env variable not defined")
		return
	}
	url, err := url.Parse(rawURL)
	if err != nil {
		log.Fatal("error")
	}
	r.GET("/ping", gin.WrapF(NewProxy(url)))

	r.Run() // listen and serve on 0.0.0.0:8080
}
