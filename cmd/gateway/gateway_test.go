package gateway

import (
	"bytes"
	"encoding/json"
	"errors"

	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"fiufit.api.gateway/internal/auth"

	"github.com/gin-gonic/gin"
)

func TestGateway(t *testing.T) {
	gin.SetMode(gin.TestMode)
	assertStatusCode := func(t testing.TB, got, want int) {
		t.Helper()
		if got != want {
			t.Errorf("Got %d, want %d", got, want)
		}
	}

	assertBody := func(t testing.TB, got, want string) {
		t.Helper()
		if got != want {
			t.Errorf("Got %s, want %s", got, want)
		}
	}
	t.Run("Redirect a request to the gateway to another service", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "application/text")
			w.Write([]byte("reverse-proxy"))
		}))
		defer server.Close()

		url, _ := url.Parse(server.URL)
		gateway := New(func(r *gin.Engine) {
			r.POST("/test", reverseProxy(url))
		})

		w := CreateTestResponseRecorder()
		req, _ := http.NewRequest("POST", "/test", nil)
		gateway.ServeHTTP(w, req)

		assertStatusCode(t, w.Code, http.StatusOK)
		assertBody(t, w.Body.String(), "reverse-proxy")
	})

	t.Run("Receive a request to sign up a new user, authorize it and notify both client and Users service",
		func(t *testing.T) {
			usersService := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusCreated)
				w.Header().Set("Content-Type", "application/text")
			}))
			defer usersService.Close()
			usersServiceURL, _ := url.Parse(usersService.URL)
			s := AuthTestService{}
			gateway := New(func(r *gin.Engine) {
				r.POST("/users", createUser(usersServiceURL, s))
			})

			signUpData := auth.SignUpModel{
				Email: "abc@xyz.com", Username: "abc", Password: "123",
			}
			signUpDataJSON, _ := json.Marshal(signUpData)
			w := CreateTestResponseRecorder()
			req, _ := http.NewRequest("POST", "/users", bytes.NewReader(signUpDataJSON))
			gateway.ServeHTTP(w, req)

			assertStatusCode(t, w.Code, http.StatusCreated)

			userData := auth.UserModel{UID: "123", Username: "abc", Email: "abc@xyz.com"}
			authDataJSON, _ := json.Marshal(userData)
			assertBody(t, w.Body.String(), string(authDataJSON))
		})

	t.Run("Receive a request to sign up a new user and fail to create it because the password is too short, the client should receive info about the error.",
		func(t *testing.T) {
			usersService := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusCreated)
				w.Header().Set("Content-Type", "application/text")
			}))
			defer usersService.Close()
			usersServiceURL, _ := url.Parse(usersService.URL)
			s := AuthTestService{}
			gateway := New(func(r *gin.Engine) {
				r.POST("/users", createUser(usersServiceURL, s))
			})

			signUpData := auth.SignUpModel{
				Email: "abc@xyz.com", Username: "abc", Password: "12",
			}
			signUpDataJSON, _ := json.Marshal(signUpData)
			w := CreateTestResponseRecorder()
			req, _ := http.NewRequest("POST", "/users", bytes.NewReader(signUpDataJSON))
			gateway.ServeHTTP(w, req)

			assertStatusCode(t, w.Code, http.StatusConflict)

			authDataJSON, _ := json.Marshal(gin.H{"error": "too short"})
			assertBody(t, w.Body.String(), string(authDataJSON))
		})
	t.Run("Receiving an invalid body should return a response with status Bad Request", func(t *testing.T) {
		usersService := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusCreated)
			w.Header().Set("Content-Type", "application/text")
		}))
		defer usersService.Close()
		usersServiceURL, _ := url.Parse(usersService.URL)
		s := AuthTestService{}
		gateway := New(func(r *gin.Engine) {
			r.POST("/users", createUser(usersServiceURL, s))
		})

		w := CreateTestResponseRecorder()
		req, _ := http.NewRequest("POST", "/users", bytes.NewReader([]byte("not json")))
		gateway.ServeHTTP(w, req)

		assertStatusCode(t, w.Code, http.StatusBadRequest)
	})

	t.Run("If the client tries to create an user and the Users service is unreachable the gateway returns Bad Gateway", func(t *testing.T) {
		url, _ := url.Parse("http://localhost:0")
		s := AuthTestService{}
		gateway := New(func(r *gin.Engine) {
			r.POST("/users", createUser(url, s))
		})

		signUpData := auth.SignUpModel{
			Email: "abc@xyz.com", Username: "abc", Password: "123",
		}
		signUpDataJSON, _ := json.Marshal(signUpData)
		w := CreateTestResponseRecorder()
		req, _ := http.NewRequest("POST", "/users", bytes.NewReader(signUpDataJSON))
		gateway.ServeHTTP(w, req)

		assertStatusCode(t, w.Code, http.StatusBadGateway)
	})
}

type AuthTestService struct{}

func (a AuthTestService) CreateUser(s auth.SignUpModel) (auth.UserModel, error) {
	if len(s.Password) < 3 {
		return auth.UserModel{}, errors.New("too short")
	}
	return auth.UserModel{UID: "123", Username: "abc", Email: "abc@xyz.com"}, nil
}

// The types below are necessary for tests to run Gin requires that
// the recorder implements the CloseNotify interface. So we generated
// a wrapper that implements it.
type TestResponseRecorder struct {
	*httptest.ResponseRecorder
	closeChannel chan bool
}

func (r *TestResponseRecorder) CloseNotify() <-chan bool {
	return r.closeChannel
}

func (r *TestResponseRecorder) closeClient() {
	r.closeChannel <- true
}

func CreateTestResponseRecorder() *TestResponseRecorder {
	return &TestResponseRecorder{
		httptest.NewRecorder(),
		make(chan bool, 1),
	}
}
