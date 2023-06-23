package gateway

import (
	"net/http"
	"net/url"

	"fiufit.api.gateway/cmd/middleware"
	"fiufit.api.gateway/internal/auth"
	"github.com/gin-gonic/gin"

	gintrace "gopkg.in/DataDog/dd-trace-go.v1/contrib/gin-gonic/gin"
)

type RouterConfig func(*gin.Engine)

type Gateway struct {
	router *gin.Engine
}

func (g *Gateway) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	g.router.ServeHTTP(w, r)
}

func (g *Gateway) Run(addr ...string) {
	g.router.Run(addr...)
}

func New(configs ...RouterConfig) *Gateway {
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.Logger())
	router.Use(gintrace.Middleware("service-external-gateway"))
	router.Use(middleware.Cors())
	for _, option := range configs {
		option(router)
	}
	return &Gateway{router}
}

// Sets the routes for the users endpoint
func Users(url *url.URL, s auth.Service) RouterConfig {
	return func(router *gin.Engine) {

		router.POST("/users",
			middleware.CreateUser(s),
			middleware.ReverseProxy(&*url))

		router.GET("/users/:user_id",
			middleware.AuthorizeUser(s),
			// Verify that it's the same user, need to notify if it's
			// owner or not
			// ?owns=true/false
			middleware.AbortIfNotAuthorized,
			middleware.ReverseProxy(&*url))

		router.GET("/users",
			middleware.AuthorizeUser(s),
			middleware.AbortIfNotAuthorized,
			middleware.SetQuery("admin", "false"),
			middleware.ReverseProxy(&*url))

		router.PUT("/users/:user_id",
			middleware.AuthorizeUser(s),
			// Verify that it's the same user,
			middleware.AbortIfNotAuthorized,
			middleware.ReverseProxy(&*url))

		router.POST("/users/:user_id/followers/:follower_id",
			middleware.AuthorizeUser(s),
			// verify that user_id is the same as the one in the token
			middleware.AbortIfNotAuthorized,
			middleware.ReverseProxy(&*url))

		router.DELETE("/users/:user_id/followers/:follower_id",
			middleware.AuthorizeUser(s),
			// verify that user_id is the same as the one in the token
			middleware.AbortIfNotAuthorized,
			middleware.ReverseProxy(&*url))

		router.GET("/users/:user_id/followers",
			middleware.ReverseProxy(&*url))

		router.GET("/users/:user_id/following",
			middleware.ReverseProxy(&*url))

		router.GET("/trainingtypes", middleware.ReverseProxy(&*url))

		router.POST("/certificates/:user_id",
			middleware.AuthorizeUser(s),
			// verify that user_id is the same as the one in the token
			middleware.AbortIfNotAuthorized,
			middleware.ReverseProxy(&*url))
		router.GET("/certificates/:user_id",
			middleware.AuthorizeUser(s),
			// verify that user_id is the same as the one in the token
			middleware.AbortIfNotAuthorized,
			middleware.ReverseProxy(&*url))

	}
}

func Admin(usersUrl *url.URL, trainersURL *url.URL, metricsURL *url.URL, s auth.Service) RouterConfig {
	return func(router *gin.Engine) {
		router.POST("/admins",
			middleware.AuthorizeUser(s),
			middleware.AbortIfNotAuthorized,
			middleware.AuthorizeAdmin(&*usersUrl),
			middleware.CreateAdmin(s),
			middleware.ReverseProxy(&*usersUrl))

		router.GET("/admins/users",
			middleware.AuthorizeUser(s),
			middleware.AbortIfNotAuthorized,
			middleware.AuthorizeAdmin(&*usersUrl),
			middleware.RemovePathFromRequestURL("/admins"),
			middleware.SetQuery("admin", "true"),
			middleware.ReverseProxy(&*usersUrl))

		// TODO: Add middleware to block in firebase
		router.PATCH("/admins/users",
			middleware.AuthorizeUser(s),
			middleware.AbortIfNotAuthorized,
			middleware.AuthorizeAdmin(&*usersUrl),
			middleware.RemovePathFromRequestURL("/admins"),
			middleware.ChangeBlockStatusFirebase(s),
			middleware.ReverseProxy(&*usersUrl))

		router.GET("/admins/plans",
			middleware.AuthorizeUser(s),
			middleware.AbortIfNotAuthorized,
			middleware.AuthorizeAdmin(&*usersUrl),
			middleware.RemovePathFromRequestURL("/admins"),
			middleware.SetQuery("admin", "true"),
			middleware.ReverseProxy(&*trainersURL))

		router.GET("/admins/plans/:trainer_id",
			middleware.AuthorizeUser(s),
			middleware.AbortIfNotAuthorized,
			middleware.AuthorizeAdmin(&*usersUrl),
			middleware.RemovePathFromRequestURL("/admins"),
			middleware.SetQuery("admin", "true"),
			middleware.ReverseProxy(&*trainersURL))

		router.PATCH("/admins/plans",
			middleware.AuthorizeUser(s),
			middleware.AbortIfNotAuthorized,
			middleware.AuthorizeAdmin(&*usersUrl),
			middleware.SetQuery("admin", "true"),
			middleware.RemovePathFromRequestURL("/admins"),
			middleware.ReverseProxy(&*trainersURL))

		router.GET("/admins/certificates",
			middleware.AuthorizeUser(s),
			middleware.AbortIfNotAuthorized,
			middleware.AuthorizeAdmin(&*usersUrl),
			middleware.RemovePathFromRequestURL("/admins"),
			middleware.ReverseProxy(&*usersUrl))

		router.PUT("/admins/certificates/:user_id/:id",
			middleware.AuthorizeUser(s),
			middleware.AbortIfNotAuthorized,
			middleware.AuthorizeAdmin(&*usersUrl),
			middleware.RemovePathFromRequestURL("/admins"),
			middleware.ReverseProxy(&*usersUrl))

		router.POST("/admins/metrics",
			middleware.AuthorizeUser(s),
			middleware.AbortIfNotAuthorized,
			middleware.AuthorizeAdmin(&*usersUrl),
			middleware.RemovePathFromRequestURL("/admins"),
			middleware.ReverseProxy(&*metricsURL))
		router.GET("/admins/metrics",
			middleware.AuthorizeUser(s),
			middleware.AbortIfNotAuthorized,
			middleware.AuthorizeAdmin(&*usersUrl),
			middleware.RemovePathFromRequestURL("/admins"),
			middleware.ReverseProxy(&*metricsURL))
		router.GET("/admins/metrics/totals",
			middleware.AuthorizeUser(s),
			middleware.AbortIfNotAuthorized,
			middleware.AuthorizeAdmin(&*usersUrl),
			middleware.RemovePathFromRequestURL("/admins"),
			middleware.ReverseProxy(&*metricsURL))
	}
}

func Trainings(url *url.URL, s auth.Service) RouterConfig {
	return func(router *gin.Engine) {
		router.POST("/plans",
			middleware.AuthorizeUser(s),
			// Verify that the user is indeed a trainer
			middleware.AbortIfNotAuthorized,
			middleware.ReverseProxy(&*url))

		router.GET("/plans",
			middleware.AuthorizeUser(s),
			middleware.AbortIfNotAuthorized,
			middleware.SetQuery("admin", "false"),
			middleware.ReverseProxy(&*url))

		router.PUT("/plans/:plan_id",
			middleware.AuthorizeUser(s),
			middleware.AbortIfNotAuthorized,
			middleware.ReverseProxy(&*url))

		router.GET("/plans/:plan_id",
			middleware.AuthorizeUser(s),
			// Verify that the user is indeed a trainer, and that it's the same
			middleware.AbortIfNotAuthorized,
			middleware.SetQuery("admin", "false"),
			middleware.ReverseProxy(&*url))

		router.GET("/trainers/:trainer_id/plans",
			middleware.AuthorizeUser(s),
			// Verify that the user is indeed a trainer, and that it's the same
			middleware.AbortIfNotAuthorized,
			middleware.SetQuery("admin", "false"),
			middleware.ReverseProxy(&*url))

		// TODO: Verify trainer_id vs token
		router.DELETE("/plans/:trainer_id/:plan_id",
			middleware.AuthorizeUser(s),
			middleware.AbortIfNotAuthorized,
			middleware.ReverseProxy(&*url))

		router.POST("/users/:user_id/trainings/favourites",
			middleware.AuthorizeUser(s),
			middleware.AbortIfNotAuthorized,
			middleware.ReverseProxy(&*url))

		router.GET("/users/:user_id/trainings/favourites",
			middleware.AuthorizeUser(s),
			middleware.AbortIfNotAuthorized,
			middleware.ReverseProxy(&*url))

		router.DELETE("/users/:user_id/trainings/favourites/:plan_id",
			middleware.AuthorizeUser(s),
			middleware.AbortIfNotAuthorized,
			middleware.ReverseProxy(&*url))
	}
}

func Reviews(url *url.URL, s auth.Service) RouterConfig {
	return func(router *gin.Engine) {
		router.POST("/reviews",
			middleware.AuthorizeUser(s),
			middleware.AbortIfNotAuthorized,
			middleware.ReverseProxy(&*url))

		router.GET("/reviews/:plan_id",
			middleware.AuthorizeUser(s),
			middleware.AbortIfNotAuthorized,
			middleware.ReverseProxy(&*url))

		router.GET("/reviews/:plan_id/mean", middleware.ReverseProxy(&*url))
	}
}

func Goals(url *url.URL, s auth.Service) RouterConfig {
	return func(router *gin.Engine) {
		router.POST("/users/:user_id/goals",
			middleware.AuthorizeUser(s),
			// Verify that is the same
			middleware.AbortIfNotAuthorized,
			middleware.ReverseProxy(&*url))

		router.PUT("/users/:user_id/goals",
			middleware.AuthorizeUser(s),
			// Verify that is the same
			middleware.AbortIfNotAuthorized,
			middleware.ReverseProxy(&*url))

		router.GET("/users/:user_id/goals",
			middleware.AuthorizeUser(s),
			// Verify that is the same
			middleware.AbortIfNotAuthorized,
			middleware.ReverseProxy(&*url))

		router.POST("/users/:user_id/training",
			middleware.AuthorizeUser(s),
			// Verify that is the same
			middleware.AbortIfNotAuthorized,
			middleware.ReverseProxy(*&url))

		router.GET("/users/:user_id/training",
			middleware.AuthorizeUser(s),
			// Verify that is the same
			middleware.AbortIfNotAuthorized,
			middleware.ReverseProxy(*&url))

		router.GET("/users/:user_id/training/metrics",
			middleware.AuthorizeUser(s),
			// Verify that is the same
			middleware.AbortIfNotAuthorized,
			middleware.ReverseProxy(*&url))
	}
}
