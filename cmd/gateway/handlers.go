package gateway

import (
	"net/url"

	"fiufit.api.gateway/cmd/middleware"
	"fiufit.api.gateway/internal/auth"
	"github.com/gin-gonic/gin"
)

func CreateUser(url *url.URL, s auth.Service) gin.HandlerFunc {
	usersServiceURL := &*url
	return func(ctx *gin.Context) {
		middleware.CreateUser(s)(ctx)
		middleware.ReverseProxy(usersServiceURL)(ctx)
	}
}

func GetAuthorizedUserProfile(url *url.URL, s auth.Service) gin.HandlerFunc {
	usersServiceURL := &*url
	return func(ctx *gin.Context) {
		middleware.AuthorizeUser(s)(ctx)
		middleware.AddUIDToRequestURL()(ctx)
		middleware.ReverseProxy(usersServiceURL)(ctx)
	}
}

func UpdateUserProfile(url *url.URL, s auth.Service) gin.HandlerFunc {
	usersServiceURL := &*url
	return func(ctx *gin.Context) {
		middleware.AuthorizeUser(s)(ctx)
		middleware.AddUIDToRequestURL()(ctx)
		middleware.ReverseProxy(usersServiceURL)(ctx)
	}
}

func CreateAdmin(url *url.URL, s auth.Service) gin.HandlerFunc {
	usersServiceURL := &*url
	return func(ctx *gin.Context) {
		middleware.AuthorizeUser(s)(ctx)
		middleware.AuthorizeAdmin(usersServiceURL)(ctx)
		middleware.CreateAdmin(s)(ctx)
		middleware.ReverseProxy(usersServiceURL)(ctx)
	}
}

func GetAllUserProfiles(url *url.URL, s auth.Service) gin.HandlerFunc {
	usersServiceURL := &*url
	return func(ctx *gin.Context) {
		middleware.AuthorizeUser(s)(ctx)
		middleware.AuthorizeAdmin(usersServiceURL)(ctx)
		middleware.RemovePathFromRequestURL("/admins")(ctx)
		middleware.ReverseProxy(usersServiceURL)(ctx)
	}
}
