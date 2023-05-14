package gateway

import (
	"net/url"

	"fiufit.api.gateway/cmd/middleware"
	"fiufit.api.gateway/internal/auth"
	"github.com/gin-gonic/gin"
)

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
