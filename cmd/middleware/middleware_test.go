package middleware

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

func assertString(t testing.TB, got, want string) {
	t.Helper()
	if got != want {
		t.Errorf("Got %s, want %s", got, want)
	}
}

func assertInt(t testing.TB, got, want int) {
	t.Helper()
	if got != want {
		t.Errorf("Got %d, want %d", got, want)
	}
}

func TestAuthorize(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Run("Authorize request correctly, the next middlware can extract the UID", func(t *testing.T) {
		s := &AuthTestService{}
		w := CreateTestResponseRecorder()
		c, _ := gin.CreateTestContext(w)
		req, _ := http.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "abc")
		c.Request = req

		AuthorizeUser(s)(c)

		anyUID, found := c.Get("User-UID")
		if !found {
			t.Errorf("Key %s wasn't found", "User-UID")
		} else {
			UID, _ := anyUID.(string)
			assertString(t, UID, "123")
		}
	})

	t.Run("Authorization of request fails, the UID isn't set. The status of the response is Unauthorized", func(t *testing.T) {
		s := &AuthTestService{}
		w := CreateTestResponseRecorder()
		c, _ := gin.CreateTestContext(w)
		req, _ := http.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "xyz")
		c.Request = req
		AuthorizeUser(s)(c)

		c.Writer.WriteHeaderNow()
		got := w.Result().StatusCode
		assertInt(t, got, http.StatusUnauthorized)

		_, found := c.Get("User-UID")
		if found {
			t.Errorf("Key %s was found", "User-UID")
		}
	})
}

func TestAddUIDToRequestURL(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Run("Add UID to URL correctly", func(t *testing.T) {
		w := CreateTestResponseRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set(uidKey, "xyz")
		url, _ := url.Parse("http://www.example.com")
		req, _ := http.NewRequest(http.MethodGet, "/test", nil)
		c.Request = req
		c.Request.URL = url
		AddUIDToRequestURL()(c)

		path := c.Request.URL.String()
		assertString(t, path, "http://www.example.com/xyz")
	})

	t.Run("Key not found so the middleware aborts with status Internal Server Error", func(t *testing.T) {
		w := CreateTestResponseRecorder()
		c, _ := gin.CreateTestContext(w)
		url, _ := url.Parse("http://www.example.com")
		req, _ := http.NewRequest(http.MethodGet, "/test", nil)
		c.Request = req
		c.Request.URL = url
		AddUIDToRequestURL()(c)
		if !c.IsAborted() {
			t.Error("The middleware didn't abort")
		}
		c.Writer.WriteHeaderNow()
		got := w.Result().StatusCode
		assertInt(t, got, http.StatusInternalServerError)
	})

	t.Run("Key couldn't be cast to string so the middleware aborts with status Internal Server Error", func(t *testing.T) {
		w := CreateTestResponseRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set(uidKey, make(chan int))
		url, _ := url.Parse("http://www.example.com")
		req, _ := http.NewRequest(http.MethodGet, "/test", nil)
		c.Request = req
		c.Request.URL = url
		AddUIDToRequestURL()(c)
		if !c.IsAborted() {
			t.Error("The middleware didn't abort")
		}
		c.Writer.WriteHeaderNow()
		got := w.Result().StatusCode
		assertInt(t, got, http.StatusInternalServerError)
	})
}

func TestReverseProxy(t *testing.T) {
	gin.SetMode(gin.TestMode)
	// Search for an easier way to test it, or remove the test altogether since
	// ReverseProxy is a wrapper over the std reverse proxy
	t.Run("Redirect a request to another service", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assertString(t, r.URL.String(), "/test")
			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", "application/text")
			w.Write([]byte("reverse-proxy"))
		}))
		defer server.Close()

		url, _ := url.Parse(server.URL)
		w := CreateTestResponseRecorder()
		_, r := gin.CreateTestContext(w)
		r.GET("/test", ReverseProxy(url))
		req, _ := http.NewRequest(http.MethodGet, "/test", nil)

		r.ServeHTTP(w, req)

		assertString(t, w.Body.String(), "reverse-proxy")
	})
}

func TestCreateUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Run("Send valid sign up data, create an user and put the user data in the request body", func(t *testing.T) {
		w := CreateTestResponseRecorder()
		c, _ := gin.CreateTestContext(w)

		s := &AuthTestService{}

		signUpData := auth.SignUpModel{
			Email: "abc@xyz.com", Username: "abc", Password: "123",
		}
		signUpDataJSON, _ := json.Marshal(signUpData)
		bodyBytes := bytes.NewReader(signUpDataJSON)
		req, _ := http.NewRequest(http.MethodPost, "/test", bodyBytes)

		c.Request = req

		CreateUser(s)(c)

		assertInt(t, s.CreateUserCalls, 1)
		body, _ := io.ReadAll(c.Request.Body)
		defer c.Request.Body.Close()

		userData := auth.UserModel{UID: "123", Username: "abc", Email: "abc@xyz.com"}
		userDataJSON, _ := json.Marshal(userData)

		assertString(t, string(body), string(userDataJSON))
	})

	// TODO: Missing testing that the body contains some context of the error
	t.Run("If the body contains invalid JSON the middleware aborts and sets the response status to Bad Request", func(t *testing.T) {
		w := CreateTestResponseRecorder()
		c, _ := gin.CreateTestContext(w)

		s := &AuthTestService{}
		req, _ := http.NewRequest(http.MethodGet, "/test", nil)
		c.Request = req
		CreateUser(s)(c)

		if !c.IsAborted() {
			t.Error("The middleware didn't abort")
		}

		c.Writer.WriteHeaderNow()
		got := w.Result().StatusCode
		assertInt(t, got, http.StatusBadRequest)
	})
	// TODO: Missing testing that the body contains some context of the error
	t.Run("If the sign up data in the body is invalid the middleware aborts and sets the response status to Conflict", func(t *testing.T) {
		w := CreateTestResponseRecorder()
		c, _ := gin.CreateTestContext(w)

		s := &AuthTestService{}
		req, _ := http.NewRequest(http.MethodGet, "/test", nil)
		c.Request = req
		CreateUser(s)(c)

		if !c.IsAborted() {
			t.Error("The middleware didn't abort")
		}

		c.Writer.WriteHeaderNow()
		got := w.Result().StatusCode
		assertInt(t, got, http.StatusBadRequest)
	})
}

func TestAuthorizeAdmin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Run("Authenticate a request from the admin", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assertString(t, r.URL.String(), "/admin/xyz")
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		url, _ := url.Parse(server.URL)
		w := CreateTestResponseRecorder()
		c, r := gin.CreateTestContext(w)
		c.Set(uidKey, "xyz")
		req, _ := http.NewRequest(http.MethodPost, "/test", nil)
		c.Request = req

		AuthorizeAdmin(url)(c)

		r.ServeHTTP(w, req)

		if c.IsAborted() {
			t.Error("The middleware shouldn't abort")
		}
	})

	t.Run("Authenticate a request from the admin, UID was not set, middleware aborts", func(t *testing.T) {
		w := CreateTestResponseRecorder()
		c, r := gin.CreateTestContext(w)
		req, _ := http.NewRequest(http.MethodPost, "/test", nil)
		c.Request = req

		AuthorizeAdmin(&url.URL{})(c)

		r.ServeHTTP(w, req)

		if !c.IsAborted() {
			t.Error("The middleware should abort")
		}
		c.Writer.WriteHeaderNow()
		got := w.Result().StatusCode
		assertInt(t, got, http.StatusInternalServerError)
	})

	t.Run("Authenticate a request from the admin, UID was set, but wasn't valid", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assertString(t, r.URL.String(), "/admin/xyz")
			w.WriteHeader(http.StatusUnauthorized)
		}))
		defer server.Close()

		url, _ := url.Parse(server.URL)
		w := CreateTestResponseRecorder()
		c, r := gin.CreateTestContext(w)
		c.Set(uidKey, "xyz")
		req, _ := http.NewRequest(http.MethodPost, "/test", nil)
		c.Request = req

		AuthorizeAdmin(url)(c)

		r.ServeHTTP(w, req)

		if !c.IsAborted() {
			t.Error("The middleware should abort")
		}

		got := w.Code
		assertInt(t, got, http.StatusUnauthorized)
	})
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

type AuthTestService struct {
	CreateUserCalls int
}

func (a *AuthTestService) CreateUser(s auth.SignUpModel) (auth.UserModel, error) {
	if len(s.Password) < 3 {
		return auth.UserModel{}, errors.New("too short")
	}
	a.CreateUserCalls += 1
	return auth.UserModel{UID: "123", Username: "abc", Email: "abc@xyz.com"}, nil
}

func (a *AuthTestService) VerifyToken(token string) (string, error) {
	if token != "abc" {
		return "", errors.New("unauthorized")
	}
	return "123", nil
}
