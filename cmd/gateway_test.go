package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"

	"testing"

	"github.com/gin-gonic/gin"
)

func TestUserSignUp(t *testing.T) {
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
		gateway := NewGateway(func(p Proxy, r *gin.Engine) {
			r.POST("/test", p(url))
		})

		w := CreateTestResponseRecorder()
		req, _ := http.NewRequest("POST", "/test", nil)

		gateway.ServeHTTP(w, req)

		assertStatusCode(t, w.Code, http.StatusOK)
		assertBody(t, w.Body.String(), "reverse-proxy")
	})

	// t.Run("Receive a request to sign up a new user, authorize it and send it to user service",
	// 	func(t *testing.T) {
	// 		testURL, _ := url.Parse("http://localhost:8080")
	// 		gateway := NewGateway(testURL)

	// 		w := httptest.NewRecorder()

	// 		body, _ := json.Marshal(SignUpData{"user@abc.com", "1234"})
	// 		req, _ := http.NewRequest("POST", "/users", bytes.NewReader(body))
	// 		gateway.ServeHTTP(w, req)
	// 		assertStatusCode(t, w.Code, http.StatusOK)

	// 		expectedJsonResponse, _ := json.Marshal(AuthData{"123", "abc", "xyz"})
	// 		assertBody(t, w.Body.String(), string(expectedJsonResponse))
	// 	})
}

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
