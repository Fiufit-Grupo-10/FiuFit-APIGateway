package main

import (
	"bytes"
	"encoding/json"

	"fiufit.api.gateway/internal/auth"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

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
		gateway := NewGateway(func(r *gin.Engine) {
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
			gateway := NewGateway(func(r *gin.Engine) {
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

			authData := auth.AuthorizedModel{UID: "123", Token: "abc", RefreshToken: "xyz"}
			authDataJSON, _ := json.Marshal(authData)
			assertBody(t, w.Body.String(), string(authDataJSON))
		})
}

type AuthTestService struct{}

func (a AuthTestService) CreateUser(s auth.SignUpModel) (auth.AuthorizedModel, error) {
	return auth.AuthorizedModel{UID: "123", Token: "abc", RefreshToken: "xyz"}, nil
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
