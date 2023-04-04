package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestUserSignUp(t *testing.T) {
	testURL, _ := url.Parse("http://localhost:8080")
	gateway := NewGateway(testURL)

	w := httptest.NewRecorder()

	body, _ := json.Marshal(SignUpData{"user@abc.com", "1234"})
	req, _ := http.NewRequest("POST", "/users", bytes.NewReader(body))
	gateway.ServeHTTP(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Got %d, want %d", w.Code, http.StatusOK)
	}

	expectedJsonResponse, _ := json.Marshal(AuthData{"123", "abc", "xyz"})
	if w.Body.String() !=  string(expectedJsonResponse){
		t.Errorf("Got %s, want %s", w.Body.String(), string(expectedJsonResponse))
	}
}
