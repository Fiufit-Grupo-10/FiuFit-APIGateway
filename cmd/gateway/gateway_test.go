package gateway

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
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

	assertString := func(t testing.TB, got, want string) {
		t.Helper()
		if got != want {
			t.Errorf("Got %s, want %s", got, want)
		}
	}

	t.Run("Receive a request to sign up a new user, authorize it and notify both client and Users service",
		func(t *testing.T) {
			usersService := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusCreated)
				w.Header().Set("Content-Type", "application/text")
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

			userData := auth.UserModel{UID: "123", Username: "abc", Email: "abc@xyz.com"}
			authDataJSON, _ := json.Marshal(userData)
			assertString(t, w.Body.String(), string(authDataJSON))
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
			gateway := New(Users(usersServiceURL, s))
			signUpData := auth.SignUpModel{
				Email: "abc@xyz.com", Username: "abc", Password: "12",
			}
			signUpDataJSON, _ := json.Marshal(signUpData)
			w := CreateTestResponseRecorder()
			req, _ := http.NewRequest("POST", "/users", bytes.NewReader(signUpDataJSON))
			gateway.ServeHTTP(w, req)

			assertStatusCode(t, w.Code, http.StatusConflict)

			authDataJSON, _ := json.Marshal(gin.H{"error": "too short"})
			assertString(t, w.Body.String(), string(authDataJSON))
		})

	t.Run("Receiving an invalid body should return a response with status Bad Request", func(t *testing.T) {
		usersService := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusCreated)
			w.Header().Set("Content-Type", "application/text")
		}))
		defer usersService.Close()
		usersServiceURL, _ := url.Parse(usersService.URL)
		s := AuthTestService{}
		gateway := New(Users(usersServiceURL, s))
		w := CreateTestResponseRecorder()
		req, _ := http.NewRequest("POST", "/users", bytes.NewReader([]byte("not json")))
		gateway.ServeHTTP(w, req)

		assertStatusCode(t, w.Code, http.StatusBadRequest)
	})

	t.Run("Receive a request to update the profile of an authorized user and forward the request to the users service",
		func(t *testing.T) {
			usersService := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				got := r.Header.Get("Authorization")
				want := "abc"
				if got != want {
					t.Errorf("Got %s, want %s", got, want)
				}
				assertString(t, r.URL.Path, "/profiles/123")
				w.WriteHeader(http.StatusOK)
				body, _ := io.ReadAll(r.Body)
				defer r.Body.Close()
				w.Write(body)
			}))
			defer usersService.Close()
			usersServiceURL, _ := url.Parse(usersService.URL)

			s := AuthTestService{}
			gateway := New(Profiles(usersServiceURL, s))
			profileData := struct {
				Data int
			}{Data: 1}
			profileDataJSON, _ := json.Marshal(profileData)

			w := CreateTestResponseRecorder()
			req, _ := http.NewRequest(http.MethodPut, "/profiles", bytes.NewReader(profileDataJSON))
			req.Header.Set("Authorization", "abc")

			gateway.ServeHTTP(w, req)

			assertStatusCode(t, w.Code, http.StatusOK)
			assertString(t, w.Body.String(), string(profileDataJSON))
		})

	t.Run("Receive a request to update the profile of an unauthorized user and respond with status Unauthorized",
		func(t *testing.T) {
			usersService := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
			defer usersService.Close()
			usersServiceURL, _ := url.Parse(usersService.URL)

			s := AuthTestService{}
			gateway := New(Profiles(usersServiceURL, s))

			profileData := struct{ Data int }{Data: 1}
			profileDataJSON, _ := json.Marshal(profileData)
			w := CreateTestResponseRecorder()
			req, _ := http.NewRequest(http.MethodPut, "/profiles", bytes.NewReader(profileDataJSON))
			req.Header.Set("Authorization", "xyz")
			gateway.ServeHTTP(w, req)

			assertStatusCode(t, w.Code, http.StatusUnauthorized)
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
