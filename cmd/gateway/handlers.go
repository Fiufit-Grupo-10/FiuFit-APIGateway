package gateway

import (
	"net/url"

	"fiufit.api.gateway/cmd/middleware"
	"fiufit.api.gateway/internal/auth"
	"github.com/gin-gonic/gin"
)

// CreateUser godoc
// @Summary      Create an user
// @Description  It first creates and verifies the new user with the authentication/authorization service. Then it forwards the necessary data to the backend. 
// @Tags         accounts
// @Accept       json
// @Produce      json
// @Success      200  {object}  
// {
//   "email": "user@mail.com",
//   "username": "user",
//   "birthday": "2000-12-21",
//   "level": "amateur",
//   "latitude": 1000,
//   "longitude": 1000,
//   "height": 180,
//   "weight": 80,
//   "gender": "M",
//   "target": "loss fat",
//   "trainingtypes": [
//     "Cardio"
//   ],
//   "user_type": "athlete"
// }
// @Failure      400  {object}  httputil.HTTPError
// @Failure      409  {object}  httputil.HTTPError
// @Failure      422  {object}  httputil.HTTPError
// @Failure      500  {object}  httputil.HTTPError
// @Router       /users [post]
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
