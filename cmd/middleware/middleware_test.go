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

func assert_eq[T comparable](t testing.TB, got, want T) {
	t.Helper()
	if got != want {
		t.Errorf("Got %v, want %v", got, want)
	}
}

func TestAuthorize(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Run("Authorize request correctly, the next middlware can extract the UID and verify that the user is authorized, Mustn't abort", func(t *testing.T) {
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
			return
		}
		UID, _ := anyUID.(string)
		assert_eq(t, UID, "123")

		anyAuthorized, found := c.Get("Authorized")
		if !found {
			t.Errorf("Key %s wasn't found", "Authorized")
		}
		authorized, _ := anyAuthorized.(bool)
		assert_eq(t, authorized, true)

		assert_eq(t, c.IsAborted(), false)
	})

	t.Run("Authorization of request fails, the UID isn't set and the next middleware can verify that the user is unauthorized . Mustn't abort", func(t *testing.T) {
		s := &AuthTestService{}
		w := CreateTestResponseRecorder()
		c, _ := gin.CreateTestContext(w)
		req, _ := http.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Authorization", "xyz")
		c.Request = req
		AuthorizeUser(s)(c)

		_, found := c.Get("User-UID")
		if found {
			t.Errorf("Key %s was found", "User-UID")
		}

		anyAuthorized, found := c.Get("Authorized")
		if !found {
			t.Errorf("Key %s wasn't found", "Authorized")
		}
		authorized, _ := anyAuthorized.(bool)
		assert_eq(t, authorized, false)
		assert_eq(t, c.IsAborted(), false)
	})
}

func TestSetQuery(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Run("Set query paramater", func(t *testing.T) {
		w := CreateTestResponseRecorder()
		c, _ := gin.CreateTestContext(w)
		req, _ := http.NewRequest(http.MethodGet, "www.example.com/test?value=1", nil)
		c.Request = req

		SetQuery("test", "query")(c)
		param := c.Query("test")

		assert_eq(t, param, "query")
		assert_eq(t, c.Query("value"), "1")
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
		assert_eq(t, path, "http://www.example.com/xyz")
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
		assert_eq(t, got, http.StatusInternalServerError)
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
		assert_eq(t, got, http.StatusInternalServerError)
	})
}

func TestReverseProxy(t *testing.T) {
	gin.SetMode(gin.TestMode)
	// Search for an easier way to test it, or remove the test altogether since
	// ReverseProxy is a wrapper over the std reverse proxy
	t.Run("Redirect a request to another service", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert_eq(t, r.URL.String(), "/test")
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

		assert_eq(t, w.Body.String(), "reverse-proxy")
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

		assert_eq(t, s.CreateUserCalls, 1)
		body, _ := io.ReadAll(c.Request.Body)
		defer c.Request.Body.Close()

		userData := auth.UserModel{UID: "123", Username: "abc", Email: "abc@xyz.com"}
		userDataJSON, _ := json.Marshal(userData)

		assert_eq(t, string(body), string(userDataJSON))
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
		assert_eq(t, got, http.StatusBadRequest)
	})
	// TODO: Missing testing that the body contains some context of the error
	t.Run("If the sign up data in the body is invalid the middleware aborts and sets the response status to Conflict", func(t *testing.T) {
		w := CreateTestResponseRecorder()
		c, _ := gin.CreateTestContext(w)

		s := &AuthTestService{}
		signUpData := auth.SignUpModel{
			Email: "abc@xyz.com", Username: "abc", Password: "1",
		}
		signUpDataJSON, _ := json.Marshal(signUpData)
		bodyBytes := bytes.NewReader(signUpDataJSON)
		req, _ := http.NewRequest(http.MethodGet, "/test", bodyBytes)
		c.Request = req
		CreateUser(s)(c)

		if !c.IsAborted() {
			t.Error("The middleware didn't abort")
		}

		c.Writer.WriteHeaderNow()
		got := w.Result().StatusCode
		assert_eq(t, got, http.StatusConflict)
	})
}

func TestAuthorizeAdmin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Run("Authenticate a request from the admin", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert_eq(t, r.URL.String(), "/admins/xyz")
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
		assert_eq(t, got, http.StatusInternalServerError)
	})

	t.Run("Authenticate a request from the admin, UID was set, but wasn't valid", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert_eq(t, r.URL.String(), "/admins/xyz")
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
		assert_eq(t, got, http.StatusUnauthorized)
	})
}

func TestCORS(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Run("The middleware should always set Access-Control-Allow-Origin and Credentials headers in the response", func(t *testing.T) {
		for _, method := range []string{http.MethodGet, http.MethodPut, http.MethodPost, http.MethodDelete, http.MethodOptions} {
			w := CreateTestResponseRecorder()
			c, _ := gin.CreateTestContext(w)
			req, _ := http.NewRequest(method, "/test", nil)
			req.Header.Set("Origin", "http://foo.bar")
			c.Request = req
			Cors()(c)
			c.Writer.WriteHeaderNow()
			assert_eq(t, w.Result().StatusCode, http.StatusOK)
			responseAllowOriginHeader := w.Result().Header.Get("Access-Control-Allow-Origin")
			responseAllowCredentialsHeader := w.Result().Header.Get("Access-Control-Allow-Credentials")

			assert_eq(t, responseAllowOriginHeader, "http://foo.bar")
			assert_eq(t, responseAllowCredentialsHeader, "true")
		}
	})
	t.Run("The middleware should alsoset Access-Control-Allow-Headers and Methods headers in the response for the OPTIONS methods", func(t *testing.T) {
		w := CreateTestResponseRecorder()
		c, _ := gin.CreateTestContext(w)
		req, _ := http.NewRequest(http.MethodOptions, "/test", nil)
		req.Header.Set("Origin", "http://foo.bar")
		c.Request = req
		Cors()(c)
		c.Writer.WriteHeaderNow()
		assert_eq(t, w.Result().StatusCode, http.StatusOK)
		responseAllowOriginHeader := w.Result().Header.Get("Access-Control-Allow-Origin")
		responseAllowCredentialsHeader := w.Result().Header.Get("Access-Control-Allow-Credentials")
		responseAllowHeadersHeader := w.Result().Header.Get("Access-Control-Allow-Headers")
		responseAllowMethodsHeader := w.Result().Header.Get("Access-Control-Allow-Methods")

		assert_eq(t, responseAllowOriginHeader, "http://foo.bar")
		assert_eq(t, responseAllowCredentialsHeader, "true")
		assert_eq(t, responseAllowHeadersHeader, allowedHeaders)
		assert_eq(t, responseAllowMethodsHeader, allowedMethods)
	})
}

func TestRemovePathFromRequestURL(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Run("Given the path /test the request URL path /test/user, after the middleware must be /users", func(t *testing.T) {
		w := CreateTestResponseRecorder()
		c, _ := gin.CreateTestContext(w)
		req, _ := http.NewRequest(http.MethodOptions, "/test/users", nil)
		c.Request = req
		path := "/test"
		RemovePathFromRequestURL(path)(c)
		assert_eq(t, c.Request.URL.Path, "/users")
	})

	t.Run("Given the path /test the request URL path /test, after the middleware must be the same", func(t *testing.T) {
		w := CreateTestResponseRecorder()
		c, _ := gin.CreateTestContext(w)
		req, _ := http.NewRequest(http.MethodOptions, "/test", nil)
		c.Request = req
		path := "/test"
		RemovePathFromRequestURL(path)(c)
		assert_eq(t, c.Request.URL.Path, "/test")
	})

	t.Run("Given the path 'test', the request URL path /test/users  after the middleware must be /users", func(t *testing.T) {
		w := CreateTestResponseRecorder()
		c, _ := gin.CreateTestContext(w)
		req, _ := http.NewRequest(http.MethodOptions, "/test/users", nil)
		c.Request = req
		path := "test"
		RemovePathFromRequestURL(path)(c)
		assert_eq(t, c.Request.URL.Path, "/users")
	})

	t.Run("Given the path '/test', the request URL path /users  after the middleware must be /users", func(t *testing.T) {
		w := CreateTestResponseRecorder()
		c, _ := gin.CreateTestContext(w)
		req, _ := http.NewRequest(http.MethodOptions, "/users", nil)
		c.Request = req
		path := "/test"
		RemovePathFromRequestURL(path)(c)
		assert_eq(t, c.Request.URL.Path, "/users")
	})
}

func TestExecuteIf(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Run("Execute the first middleware if the guard returns true", func(t *testing.T) {
		w := CreateTestResponseRecorder()
		c, _ := gin.CreateTestContext(w)
		req, _ := http.NewRequest(http.MethodGet, "/test", nil)
		c.Request = req

		guard := func(c *gin.Context) bool {
			return true
		}

		first := func(c *gin.Context) {
			c.Set("key", "first")
		}

		second := func(c *gin.Context) {
			c.Set("key", "second")
		}

		ExecuteIf(guard, first, second)(c)
		anyKey, found := c.Get("key")
		if !found {
			t.Errorf("Key %s wasn't found", "key")
			return
		}
		key, _ := anyKey.(string)
		assert_eq(t, key, "first")

	})

	t.Run("Execute the second middleware if the guard returns false", func(t *testing.T) {
		w := CreateTestResponseRecorder()
		c, _ := gin.CreateTestContext(w)
		req, _ := http.NewRequest(http.MethodGet, "/test", nil)
		c.Request = req

		guard := func(c *gin.Context) bool {
			return false
		}

		first := func(c *gin.Context) {
			c.Set("key", "first")
		}

		second := func(c *gin.Context) {
			c.Set("key", "second")
		}

		ExecuteIf(guard, first, second)(c)
		anyKey, found := c.Get("key")
		if !found {
			t.Errorf("Key %s wasn't found", "key")
			return
		}
		key, _ := anyKey.(string)
		assert_eq(t, key, "second")
	})
}

func TestIsAuthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	t.Run("The key is set and the user is authorized, mustn't abort", func(t *testing.T) {
		w := CreateTestResponseRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set(authorizedKey, true)
		got := IsAuthorized(c)
		assert_eq(t, got, true)
		assert_eq(t, c.IsAborted(), false)
	})

	t.Run("The key is set and the user isn't authorized, mustn't abort", func(t *testing.T) {
		w := CreateTestResponseRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set(authorizedKey, false)
		got := IsAuthorized(c)
		assert_eq(t, got, false)
		assert_eq(t, c.IsAborted(), false)
	})

	t.Run("The key isnt set , must abort", func(t *testing.T) {
		w := CreateTestResponseRecorder()
		c, _ := gin.CreateTestContext(w)
		IsAuthorized(c)
		assert_eq(t, c.IsAborted(), true)
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
