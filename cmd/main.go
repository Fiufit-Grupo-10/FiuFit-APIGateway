package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"

	"github.com/gin-gonic/gin"
)

type Gateway struct{
	router *gin.Engine
}

func NewGateway(url *url.URL) *Gateway {
	router := gin.Default()
	router.POST("/users", func(ctx *gin.Context) {
		ctx.JSONP(http.StatusOK, AuthData{"123", "abc", "xyz"})
	})
	return &Gateway{router}
}

func (g *Gateway) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	g.router.ServeHTTP(w, r)
}

type SignUpData struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthData struct {
	UID string `json:"uid"`
	Token string `json:"token"`
	RefreshToken string `json:"refresh_token"`
}

func NewProxy(url *url.URL) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		proxy := httputil.NewSingleHostReverseProxy(url)
		proxy.ServeHTTP(w, r)
	}
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
