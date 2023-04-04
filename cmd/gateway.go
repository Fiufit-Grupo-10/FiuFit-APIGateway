package main

import (
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/gin-gonic/gin"
)

type RouteConfig func(Proxy, *gin.Engine)
type Proxy func(*url.URL) gin.HandlerFunc

type Gateway struct {
	router *gin.Engine
}

func (g *Gateway) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	g.router.ServeHTTP(w, r)
}

func NewGateway(configs ...RouteConfig) *Gateway {
	router := gin.Default()
	for  _, option := range configs {
		option(reverseProxy, router)
	}
	return &Gateway{router}
}

func reverseProxy(url *url.URL) gin.HandlerFunc {
	return func(c *gin.Context) {
		proxy := httputil.NewSingleHostReverseProxy(url)
		proxy.ServeHTTP(c.Writer, c.Request)
	}
}
