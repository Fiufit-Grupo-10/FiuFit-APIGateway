package gateway

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"

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

	assertString := func(t testing.TB, got, want string) {
		t.Helper()
		if got != want {
			t.Errorf("Got %s, want %s", got, want)
		}
	}

	t.Run("Receive a request to sign up a new user",
		func(t *testing.T) {
			usersService := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				userData := auth.UserModel{UID: "123", Username: "abc", Email: "abc@xyz.com"}
				userDataJSON, _ := json.Marshal(userData)
				body, _ := io.ReadAll(r.Body)
				defer r.Body.Close()
				assertString(t, string(body), string(userDataJSON))
				w.WriteHeader(http.StatusCreated)
			}))
			defer usersService.Close()

			usersServiceURL, _ := url.Parse(usersService.URL)
			s := AuthTestService{}
			gateway := New(Users(usersServiceURL, s))

			signUpData := auth.SignUpModel{
				Email: "abc@xyz.com", Username: "abc", Password: "123",
			}
			signUpDataJSON, _ := json.Marshal(signUpData)
			w := CreateTestResponseRecorder()
			req, _ := http.NewRequest("POST", "/users", bytes.NewReader(signUpDataJSON))
			gateway.ServeHTTP(w, req)
			assertStatusCode(t, w.Code, http.StatusCreated)
		})

	t.Run("Receive a request to update the profile of an authorized user and forward the request to the users service",
		func(t *testing.T) {
			profileData := struct {
				Data int
			}{Data: 1}
			profileDataJSON, _ := json.Marshal(profileData)

			usersService := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assertString(t, r.URL.Path, "/users/123")
				w.WriteHeader(http.StatusOK)
				body, _ := io.ReadAll(r.Body)
				defer r.Body.Close()
				assertString(t, string(body), string(profileDataJSON))
				w.Write(body)
			}))
			defer usersService.Close()

			usersServiceURL, _ := url.Parse(usersService.URL)

			s := AuthTestService{}
			gateway := New(Users(usersServiceURL, s))

			w := CreateTestResponseRecorder()
			req, _ := http.NewRequest(http.MethodPut, "/users/123", bytes.NewReader(profileDataJSON))
			req.Header.Set("Authorization", "abc")

			gateway.ServeHTTP(w, req)

			assertStatusCode(t, w.Code, http.StatusOK)
		})

	t.Run("Request an user profile, being authorized, the response body must be a json containing the users private profile data",
		func(t *testing.T) {
			userData := auth.UserModel{UID: "123", Username: "abc", Email: "abc@xyz.com"}
			userDataJSON, _ := json.Marshal(userData)
			usersService := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assertString(t, r.URL.Path, "/users/123")
				w.WriteHeader(http.StatusCreated)
				w.Write(userDataJSON)
			}))
			defer usersService.Close()

			usersServiceURL, _ := url.Parse(usersService.URL)
			s := AuthTestService{}
			gateway := New(Users(usersServiceURL, s))
			w := CreateTestResponseRecorder()
			req, _ := http.NewRequest(http.MethodGet, "/users", nil)
			req.Header.Set("Authorization", "abc")

			gateway.ServeHTTP(w, req)
			assertString(t, w.Body.String(), string(userDataJSON))
		})

	t.Run("Request an user profile, being unauthorized, the response body must be a json containing the users public profile data",
		func(t *testing.T) {
			userData := auth.UserModel{UID: "123", Username: "abc", Email: "abc@xyz.com"}
			userDataJSON, _ := json.Marshal(userData)
			usersService := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assertString(t, r.URL.Path, "/users")
				w.WriteHeader(http.StatusCreated)
				w.Write(userDataJSON)
			}))
			defer usersService.Close()

			usersServiceURL, _ := url.Parse(usersService.URL)
			s := AuthTestService{}
			gateway := New(Users(usersServiceURL, s))
			w := CreateTestResponseRecorder()
			req, _ := http.NewRequest(http.MethodGet, "/users", nil)
			req.Header.Set("Authorization", "xyz")

			gateway.ServeHTTP(w, req)
			assertString(t, w.Body.String(), string(userDataJSON))
		})

	t.Run("Receive a request to create a new admin, the one making the request is an admin", func(t *testing.T) {
		usersService := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodGet {
				assertString(t, r.URL.Path, "/admins/123")
			}
			w.WriteHeader(http.StatusOK)
		}))
		defer usersService.Close()

		usersServiceURL, _ := url.Parse(usersService.URL)
		s := AuthTestService{}
		gateway := New(Admin(usersServiceURL, s))

		signUpData := auth.SignUpModel{
			Email: "abc@xyz.com", Username: "abc", Password: "123",
		}
		signUpDataJSON, _ := json.Marshal(signUpData)
		w := CreateTestResponseRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/admins", bytes.NewReader(signUpDataJSON))
		req.Header.Set("Authorization", "abc")
		gateway.ServeHTTP(w, req)
	})

	t.Run("Receive a request to create a new admin, the one making the request isn't an admin, returns Unauthorized", func(t *testing.T) {
		usersService := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodGet {
				assertString(t, r.URL.Path, "/admins/123")
			}
			w.WriteHeader(http.StatusNotFound)
		}))
		defer usersService.Close()

		usersServiceURL, _ := url.Parse(usersService.URL)
		s := AuthTestService{}
		gateway := New(Admin(usersServiceURL, s))

		signUpData := auth.SignUpModel{
			Email: "abc@xyz.com", Username: "abc", Password: "123",
		}
		signUpDataJSON, _ := json.Marshal(signUpData)
		w := CreateTestResponseRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/admins", bytes.NewReader(signUpDataJSON))
		req.Header.Set("Authorization", "abc")
		gateway.ServeHTTP(w, req)
		assertStatusCode(t, w.Code, http.StatusUnauthorized)
	})

	t.Run("An admin request all the profiles", func(t *testing.T) {
		usersService := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/admins") {
				/// Verify admin
				assertString(t, r.URL.Path, "/admins/123")
			} else {
				// Gets profiles
				assertString(t, r.URL.Path, "/users")
				assertString(t, r.URL.Query().Get("admin"), "true")
			}
			w.WriteHeader(http.StatusOK)
		}))
		defer usersService.Close()

		usersServiceURL, _ := url.Parse(usersService.URL)
		s := AuthTestService{}
		gateway := New(Admin(usersServiceURL, s))
		w := CreateTestResponseRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/admins/users", nil)
		req.Header.Set("Authorization", "abc")
		gateway.ServeHTTP(w, req)
	})

	t.Run("An user request all the profiles", func(t *testing.T) {
		usersService := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assertString(t, r.URL.Path, "/users")
			assertString(t, r.URL.Query().Get("admin"), "false")
			w.WriteHeader(http.StatusOK)
		}))
		defer usersService.Close()

		usersServiceURL, _ := url.Parse(usersService.URL)
		s := AuthTestService{}
		gateway := New(Users(usersServiceURL, s))
		w := CreateTestResponseRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/users", nil)
		req.Header.Set("Authorization", "xyz")
		gateway.ServeHTTP(w, req)
	})
}

type AuthTestService struct{}

func (a AuthTestService) CreateUser(s auth.SignUpModel) (auth.UserModel, error) {
	if len(s.Password) < 3 {
		return auth.UserModel{}, errors.New("too short")
	}
	return auth.UserModel{UID: "123", Username: "abc", Email: "abc@xyz.com"}, nil
}

func (a AuthTestService) VerifyToken(token string) (string, error) {
	if token != "abc" {
		return "", errors.New("unauthorized")
	}
	return "123", nil
}

func (a AuthTestService) GetUser(uid string) (auth.UserModel, error) {
	return auth.UserModel{}, nil
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
