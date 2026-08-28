package httpx

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequireOriginAllowsExactOrigin(t *testing.T) {
	handler := RequireOrigin("https://shop.example.com")(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	request := httptest.NewRequest(http.MethodPost, "/api/auth/login", nil)
	request.Header.Set("Origin", "https://shop.example.com")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusNoContent {
		t.Fatalf("status = %d", response.Code)
	}
}

func TestRequireOriginRejectsMissingOrSiblingOrigin(t *testing.T) {
	handler := RequireOrigin("https://shop.example.com")(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	for _, origin := range []string{"", "https://evil.example.com", "https://shop.example.com.evil.test"} {
		request := httptest.NewRequest(http.MethodPost, "/api/auth/login", nil)
		request.Header.Set("Origin", origin)
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		if response.Code != http.StatusForbidden {
			t.Errorf("origin %q status = %d", origin, response.Code)
		}
	}
}
