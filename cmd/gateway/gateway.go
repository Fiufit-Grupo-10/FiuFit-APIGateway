package gateway

import (
	"net/http"
	"net/url"

	"fiufit.api.gateway/cmd/middleware"
	"fiufit.api.gateway/internal/auth"
	"fiufit.api.gateway/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/mvrilo/go-redoc"
	ginredoc "github.com/mvrilo/go-redoc/gin"
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

func New(c *config.Config ,routers ...RouterConfig) *Gateway {
	doc := redoc.Redoc{
			Title:       "FiuFit API Gateway",
			Description: "API Gateway for FiuFit App",
			SpecFile:    "./openapi.json", // "./openapi.yaml"
			SpecPath:    "/openapi.json",  // "/openapi.yaml"
			DocsPath:    "/docs",
	}
	router := gin.New()
	if !c.IsDevEnviroment {
		router.Use(ginredoc.New(doc))
	}
	router.Use(gin.Recovery())
	router.Use(middleware.Logger())

	router.Use(gintrace.Middleware("service-external-gateway"))
	router.Use(middleware.Cors())


	for _, option := range routers {
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
			middleware.ReverseProxy(&*url))

		router.GET("/users",
			middleware.AuthorizeUser(s),
			middleware.SetQuery("admin", "false"),
			middleware.ReverseProxy(&*url))

		router.PUT("/users/:user_id",
			middleware.AuthorizeUser(s),
			middleware.ReverseProxy(&*url))

		router.POST("/users/:user_id/followers/:follower_id",
			middleware.AuthorizeUser(s),
			middleware.ReverseProxy(&*url))

		router.DELETE("/users/:user_id/followers/:follower_id",
			middleware.AuthorizeUser(s),
			middleware.ReverseProxy(&*url))

		router.GET("/users/:user_id/followers",
			middleware.ReverseProxy(&*url))

		router.GET("/users/:user_id/following",
			middleware.ReverseProxy(&*url))

		router.GET("/trainingtypes", middleware.ReverseProxy(&*url))

		router.POST("/certificates/:user_id",
			middleware.AuthorizeUser(s),
			middleware.ReverseProxy(&*url))
		router.GET("/certificates/:user_id",
			middleware.AuthorizeUser(s),
			middleware.ReverseProxy(&*url))

		router.GET("/users/:user_id/trainers",
			middleware.AuthorizeUser(s),
			middleware.ReverseProxy(&*url))

	}
}

func Admin(usersUrl *url.URL, trainersURL *url.URL, metricsURL *url.URL, s auth.Service) RouterConfig {
	return func(router *gin.Engine) {
		router.POST("/admins",
			middleware.AuthorizeUser(s),
			middleware.AuthorizeAdmin(&*usersUrl),
			middleware.CreateAdmin(s),
			middleware.ReverseProxy(&*usersUrl))

		router.GET("/admins/users",
			middleware.AuthorizeUser(s),
			middleware.AuthorizeAdmin(&*usersUrl),
			middleware.RemovePathFromRequestURL("/admins"),
			middleware.SetQuery("admin", "true"),
			middleware.ReverseProxy(&*usersUrl))

		// TODO: Add middleware to block in firebase
		router.PATCH("/admins/users",
			middleware.AuthorizeUser(s),
			middleware.AuthorizeAdmin(&*usersUrl),
			middleware.RemovePathFromRequestURL("/admins"),
			middleware.ChangeBlockStatusFirebase(s),
			middleware.ReverseProxy(&*usersUrl))

		router.GET("/admins/plans",
			middleware.AuthorizeUser(s),
			middleware.AuthorizeAdmin(&*usersUrl),
			middleware.RemovePathFromRequestURL("/admins"),
			middleware.SetQuery("admin", "true"),
			middleware.ReverseProxy(&*trainersURL))

		router.GET("/admins/plans/:trainer_id",
			middleware.AuthorizeUser(s),
			middleware.AuthorizeAdmin(&*usersUrl),
			middleware.RemovePathFromRequestURL("/admins"),
			middleware.SetQuery("admin", "true"),
			middleware.ReverseProxy(&*trainersURL))

		router.PATCH("/admins/plans",
			middleware.AuthorizeUser(s),
			middleware.AuthorizeAdmin(&*usersUrl),
			middleware.SetQuery("admin", "true"),
			middleware.RemovePathFromRequestURL("/admins"),
			middleware.ReverseProxy(&*trainersURL))

		router.GET("/admins/certificates",
			middleware.AuthorizeUser(s),
			middleware.AuthorizeAdmin(&*usersUrl),
			middleware.RemovePathFromRequestURL("/admins"),
			middleware.ReverseProxy(&*usersUrl))

		router.PUT("/admins/certificates/:user_id/:id",
			middleware.AuthorizeUser(s),
			middleware.AuthorizeAdmin(&*usersUrl),
			middleware.RemovePathFromRequestURL("/admins"),
			middleware.ReverseProxy(&*usersUrl))

		router.POST("/admins/metrics",
			middleware.AuthorizeUser(s),
			middleware.RemovePathFromRequestURL("/admins"),
			middleware.ReverseProxy(&*metricsURL))

		router.GET("/admins/metrics",
			middleware.AuthorizeUser(s),
			middleware.AuthorizeAdmin(&*usersUrl),
			middleware.RemovePathFromRequestURL("/admins"),
			middleware.ReverseProxy(&*metricsURL))

		router.GET("/admins/metrics/totals",
			middleware.AuthorizeUser(s),
			middleware.AuthorizeAdmin(&*usersUrl),
			middleware.RemovePathFromRequestURL("/admins"),
			middleware.ReverseProxy(&*metricsURL))

		router.GET("/admins/metrics/locations",
			middleware.AuthorizeUser(s),
			middleware.AuthorizeAdmin(&*usersUrl),
			middleware.RemovePathFromRequestURL("/admins"),
			middleware.ReverseProxy(&*metricsURL))
	}
}

func Trainings(url *url.URL, s auth.Service) RouterConfig {
	return func(router *gin.Engine) {
		router.POST("/plans",
			middleware.AuthorizeUser(s),
			middleware.ReverseProxy(&*url))

		router.GET("/plans",
			middleware.AuthorizeUser(s),
			middleware.SetQuery("admin", "false"),
			middleware.ReverseProxy(&*url))

		router.PUT("/plans/:plan_id",
			middleware.AuthorizeUser(s),
			middleware.ReverseProxy(&*url))

		router.GET("/plans/:plan_id",
			middleware.AuthorizeUser(s),
			middleware.SetQuery("admin", "false"),
			middleware.ReverseProxy(&*url))

		router.GET("/trainers/:trainer_id/plans",
			middleware.AuthorizeUser(s),
			middleware.SetQuery("admin", "false"),
			middleware.ReverseProxy(&*url))

		router.DELETE("/plans/:trainer_id/:plan_id",
			middleware.AuthorizeUser(s),
			middleware.ReverseProxy(&*url))

		router.POST("/users/:user_id/trainings/favourites",
			middleware.AuthorizeUser(s),
			middleware.ReverseProxy(&*url))

		router.GET("/users/:user_id/trainings/favourites",
			middleware.AuthorizeUser(s),
			middleware.ReverseProxy(&*url))

		router.DELETE("/users/:user_id/trainings/favourites/:plan_id",
			middleware.AuthorizeUser(s),
			middleware.ReverseProxy(&*url))
	}
}

func Reviews(url *url.URL, s auth.Service) RouterConfig {
	return func(router *gin.Engine) {
		router.POST("/reviews",
			middleware.AuthorizeUser(s),
			middleware.ReverseProxy(&*url))

		router.GET("/reviews/:plan_id",
			middleware.AuthorizeUser(s),
			middleware.ReverseProxy(&*url))

		router.GET("/reviews/:plan_id/mean", middleware.ReverseProxy(&*url))
		router.PUT("/reviews/:review_id",
			middleware.AuthorizeUser(s),
			middleware.ReverseProxy(&*url))
	}
}

func Goals(url *url.URL, s auth.Service) RouterConfig {
	return func(router *gin.Engine) {
		router.POST("/users/:user_id/goals",
			middleware.AuthorizeUser(s),
			middleware.ReverseProxy(&*url))

		router.PUT("/users/:user_id/goals",
			middleware.AuthorizeUser(s),
			middleware.ReverseProxy(&*url))

		router.GET("/users/:user_id/goals",
			middleware.AuthorizeUser(s),
			middleware.ReverseProxy(&*url))

		router.POST("/users/:user_id/training",
			middleware.AuthorizeUser(s),
			middleware.ReverseProxy(*&url))

		router.GET("/users/:user_id/training",
			middleware.AuthorizeUser(s),
			middleware.ReverseProxy(*&url))

		router.GET("/users/:user_id/training/metrics",
			middleware.AuthorizeUser(s),
			middleware.ReverseProxy(*&url))
	}
}

func Metrics(url *url.URL, s auth.Service) RouterConfig {
	return func(router *gin.Engine) {
		router.GET("/metrics/trainings/:plan_id",
			middleware.AuthorizeUser(s),
			middleware.ReverseProxy(*&url))
	}
}
