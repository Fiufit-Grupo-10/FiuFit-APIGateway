package middleware

import (
	"errors"
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

func assertStatusCode(t testing.TB, got, want int) {
	t.Helper()
	if got != want {
		t.Errorf("Got %d, want %d", got, want)
	}
}

func TestAuthorize(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Run("Authorize request correctly, the next middlware can extract the UID", func(t *testing.T) {
		s := AuthTestService{}
		w := CreateTestResponseRecorder()
		c, _ := gin.CreateTestContext(w)
		req, _ := http.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "abc")
		c.Request = req

		Authorize(s)(c)

		anyUID, found := c.Get("User-UID")
		if !found {
			t.Errorf("Key %s wasn't found", "User-UID")
		} else {
			UID, _ := anyUID.(string)
			assertString(t, UID, "123")
		}
	})

	t.Run("Authorization of request fails, the UID isn't set", func(t *testing.T) {
		s := AuthTestService{}
		w := CreateTestResponseRecorder()
		c, _ := gin.CreateTestContext(w)
		req, _ := http.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "xyz")
		c.Request = req
		Authorize(s)(c)

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
		assertStatusCode(t, got, http.StatusInternalServerError)
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
		assertStatusCode(t, got, http.StatusInternalServerError)
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
