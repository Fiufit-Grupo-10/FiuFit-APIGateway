package gateway

import (
	"log"
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

func GetUsersProfiles(url *url.URL, s auth.Service) gin.HandlerFunc {
	usersServiceURL := &*url
	return func(ctx *gin.Context) {
		middleware.AuthorizeUser(s)(ctx)
		middleware.ExecuteIf(middleware.IsAuthorized, middleware.AddUIDToRequestURL(), middleware.SetQuery("admin", "false"))(ctx)
		middleware.ReverseProxy(usersServiceURL)(ctx)
	}
}

func UpdateUserProfile(url *url.URL, s auth.Service) gin.HandlerFunc {
	usersServiceURL := &*url
	return func(ctx *gin.Context) {
		log.Printf("========= ID: %s =========\n", ctx.Param("user_id"))
		middleware.AuthorizeUser(s)(ctx)
		// Verify that it's the same user
		middleware.AbortIfNotAuthorized(ctx)
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

func GetUsersProfilesAdmin(url *url.URL, s auth.Service) gin.HandlerFunc {
	usersServiceURL := &*url
	return func(ctx *gin.Context) {
		middleware.AuthorizeUser(s)(ctx)
		middleware.AbortIfNotAuthorized(ctx)
		middleware.AuthorizeAdmin(usersServiceURL)(ctx)
		middleware.RemovePathFromRequestURL("/admins")(ctx)
		middleware.SetQuery("admin", "true")(ctx)
		middleware.ReverseProxy(usersServiceURL)(ctx)
	}
}

func GetTrainingTypes(url *url.URL) gin.HandlerFunc {
	usersServiceURL := &*url
	return func(ctx *gin.Context) {
		middleware.ReverseProxy(usersServiceURL)(ctx)
	}
}

func CreateTrainingPlan(url *url.URL, s auth.Service) gin.HandlerFunc {
	trainersServiceURL := &*url
	return func(ctx *gin.Context) {
		middleware.AuthorizeUser(s)(ctx)
		// Verify that the user is indeed a trainer
		middleware.AbortIfNotAuthorized(ctx)
		middleware.ReverseProxy(trainersServiceURL)(ctx)
	}
}

func ModifyTrainingPlan(url *url.URL, s auth.Service) gin.HandlerFunc {
	trainersServiceURL := &*url
	return func(ctx *gin.Context) {
		middleware.AuthorizeUser(s)(ctx)
		// Verify that the user is indeed a trainer
		middleware.AbortIfNotAuthorized(ctx)
		middleware.ReverseProxy(trainersServiceURL)(ctx)
	}
}

func GetTrainerPlans(url *url.URL, s auth.Service) gin.HandlerFunc {
	trainersServiceURL := &*url
	return func(ctx *gin.Context) {
		middleware.AuthorizeUser(s)(ctx)
		// Verify that the user is indeed a trainer, and that it's the same
		middleware.AbortIfNotAuthorized(ctx)
		middleware.ReverseProxy(trainersServiceURL)(ctx)
	}
}

func GetAllTrainersPlans(url *url.URL, s auth.Service) gin.HandlerFunc {
	trainersServiceURL := &*url
	return func(ctx *gin.Context) {
		middleware.AuthorizeUser(s)(ctx)
		middleware.AbortIfNotAuthorized(ctx)
		middleware.ReverseProxy(trainersServiceURL)(ctx)
	}
}

func CreateReview(url *url.URL, s auth.Service) gin.HandlerFunc {
	trainersServiceURL := &*url
	return func(ctx *gin.Context) {
		middleware.AuthorizeUser(s)(ctx)
		middleware.AbortIfNotAuthorized(ctx)
		middleware.ReverseProxy(trainersServiceURL)(ctx)
	}
}

func GetTrainingPlanReviews(url *url.URL, s auth.Service) gin.HandlerFunc {
	trainersServiceURL := &*url
	return func(ctx *gin.Context) {
		middleware.AuthorizeUser(s)(ctx)
		// Verify that the user is indeed a trainer, and that it's the same
		middleware.AbortIfNotAuthorized(ctx)
		middleware.ReverseProxy(trainersServiceURL)(ctx)
	}
}
