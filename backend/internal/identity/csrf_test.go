package identity

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCSRFTokenIsSessionBoundAndRequired(t *testing.T) {
	handler := NewHandler(nil, slog.New(slog.NewTextHandler(io.Discard, nil)), CookieConfig{Name: "aurora_session"})
	token := csrfToken("session-a")
	if token == csrfToken("session-b") {
		t.Fatal("CSRF tokens should differ between sessions")
	}
	protected := handler.RequireCSRF(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNoContent) }))
	for _, test := range []struct {
		name, cookie, header string
		status               int
	}{
		{"valid", "session-a", token, http.StatusNoContent},
		{"missing token", "session-a", "", http.StatusForbidden},
		{"other session", "session-b", token, http.StatusForbidden},
		{"missing session", "", token, http.StatusForbidden},
	} {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/api/cart/items", strings.NewReader("{}"))
			if test.cookie != "" {
				request.AddCookie(&http.Cookie{Name: "aurora_session", Value: test.cookie})
			}
			request.Header.Set("X-CSRF-Token", test.header)
			response := httptest.NewRecorder()
			protected.ServeHTTP(response, request)
			if response.Code != test.status {
				t.Fatalf("status = %d, want %d", response.Code, test.status)
			}
		})
	}
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/auth/csrf", nil)
	request.AddCookie(&http.Cookie{Name: "aurora_session", Value: "session-a"})
	handler.CSRFToken(response, request)
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), token) || response.Header().Get("Cache-Control") != "private, no-store" {
		t.Fatalf("unexpected CSRF response: %d %q", response.Code, response.Body.String())
	}
}
