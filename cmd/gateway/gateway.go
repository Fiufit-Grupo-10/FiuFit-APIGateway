package gateway

import (
	"fiufit.api.gateway/cmd/middleware"
	"fiufit.api.gateway/internal/auth"
	"github.com/gin-gonic/gin"
	"net/http"
	"net/url"
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
	router := gin.Default()
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

		// TODO: Ask front
		router.GET("/users/:user_id",
			middleware.AuthorizeUser(s),
			middleware.ReverseProxy(&*url))

		router.GET("/users", middleware.AuthorizeUser(s),
			middleware.ExecuteIf(middleware.IsAuthorized,
				middleware.AddUIDToRequestURL(),
				middleware.SetQuery("admin", "false")),
			middleware.ReverseProxy(&*url))

		router.PUT("/users/:user_id",
			middleware.AuthorizeUser(s),
			// Verify that it's the same user
			middleware.AbortIfNotAuthorized,
			middleware.ReverseProxy(&*url))

		router.POST("/users/:user_id/followers/:follower_id", middleware.ReverseProxy(&*url))

		router.DELETE("/users/:user_id/followers/:follower_id", middleware.ReverseProxy(&*url))

		router.GET("/users/:user_id/followers", middleware.ReverseProxy(&*url))

		router.GET("/users/:user_id/following", middleware.ReverseProxy(&*url))

		router.GET("/trainingtypes", middleware.ReverseProxy(&*url))
	}
}

func Admin(usersUrl *url.URL, trainersURL *url.URL, s auth.Service) RouterConfig {
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
	}
}

func Trainers(url *url.URL, s auth.Service) RouterConfig {
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
			// Get a users
			// GET: /users?field=role
			// Verify that the user is indeed a trainer
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
			// Verify that the user is indeed a trainer, and that it's the same
			middleware.AbortIfNotAuthorized,
			middleware.ReverseProxy(&*url))

		router.GET("/reviews/:plan_id/mean", middleware.ReverseProxy(&*url))
	}
}

func Goals(url *url.URL, s auth.Service) RouterConfig {
	return func(router *gin.Engine) {
		router.POST("/users/:user_id/goals",
			middleware.AuthorizeUser(s),
			middleware.AbortIfNotAuthorized,
			middleware.ReverseProxy(&*url))

		router.PUT("/users/:user_id/goals",
			middleware.AuthorizeUser(s),
			middleware.AbortIfNotAuthorized,
			middleware.ReverseProxy(&*url))

		router.GET("/users/:user_id/goals",
			middleware.AuthorizeUser(s),
			middleware.AbortIfNotAuthorized,
			middleware.ReverseProxy(&*url))

		router.POST("/users/:user_id/training",
			middleware.AuthorizeUser(s),
			middleware.AbortIfNotAuthorized,
			middleware.ReverseProxy(*&url))

		router.GET("/users/:user_id/training",
			middleware.AuthorizeUser(s),
			middleware.AbortIfNotAuthorized,
			middleware.ReverseProxy(*&url))

		router.GET("/users/:user_id/training/metrics",
			middleware.AuthorizeUser(s),
			middleware.AbortIfNotAuthorized,
			middleware.ReverseProxy(*&url))
	}
}

// Move to admin
func Metrics(url *url.URL, s auth.Service) RouterConfig {
	return func(router *gin.Engine) {
		router.POST("/metrics", middleware.ReverseProxy(&*url))
		router.GET("/metrics", middleware.ReverseProxy(&*url))
		router.GET("/metrics/totals", middleware.ReverseProxy(&*url))
	}
}
