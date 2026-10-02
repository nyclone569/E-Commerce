package identity

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestRegisterSetsProtectedSessionCookieWithoutLeakingToken(t *testing.T) {
	repository := &stubRepository{createdUser: User{ID: uuid.New(), Email: "learner@example.com", DisplayName: "Learner", Status: "active"}}
	service := newTestService(t, repository, &stubHasher{})
	service.now = time.Now
	handler := NewHandler(service, slog.New(slog.NewTextHandler(io.Discard, nil)), CookieConfig{Name: "aurora_session", Secure: true})
	request := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBufferString(`{"email":"learner@example.com","password":"a sufficiently long passphrase","display_name":"Learner"}`))
	response := httptest.NewRecorder()

	handler.Register(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
	cookies := response.Result().Cookies()
	if len(cookies) != 1 || cookies[0].Name != "aurora_session" || !cookies[0].HttpOnly || !cookies[0].Secure || cookies[0].SameSite != http.SameSiteLaxMode || cookies[0].MaxAge <= 0 {
		t.Fatalf("unexpected session cookie: %#v", cookies)
	}
	if strings.Contains(response.Body.String(), "raw-token") || strings.Contains(response.Body.String(), "hash:") {
		t.Fatalf("response leaked credential material: %s", response.Body.String())
	}
}

func TestLoginUsesConsistentInvalidCredentialResponse(t *testing.T) {
	repository := &stubRepository{credentialErr: ErrInvalidCredentials}
	service := newTestService(t, repository, &stubHasher{})
	handler := NewHandler(service, slog.New(slog.NewTextHandler(io.Discard, nil)), CookieConfig{Name: "aurora_session"})
	request := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBufferString(`{"email":"missing@example.com","password":"submitted password"}`))
	response := httptest.NewRecorder()

	handler.Login(response, request)

	if response.Code != http.StatusUnauthorized || !strings.Contains(response.Body.String(), `"code":"invalid_credentials"`) || strings.Contains(response.Body.String(), "missing@example.com") {
		t.Fatalf("unexpected status/body: %d %s", response.Code, response.Body.String())
	}
}

func TestLogoutIsIdempotentAndClearsCookie(t *testing.T) {
	service := newTestService(t, &stubRepository{}, &stubHasher{})
	handler := NewHandler(service, slog.New(slog.NewTextHandler(io.Discard, nil)), CookieConfig{Name: "aurora_session"})
	request := httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
	response := httptest.NewRecorder()

	handler.Logout(response, request)

	if response.Code != http.StatusNoContent {
		t.Fatalf("status = %d", response.Code)
	}
	cookies := response.Result().Cookies()
	if len(cookies) != 1 || cookies[0].MaxAge != -1 {
		t.Fatalf("logout cookie = %#v", cookies)
	}
}

func TestAuthenticateDoesNotReachCartHandlerWithoutValidSession(t *testing.T) {
	repository := &stubRepository{activeErr: ErrUnauthenticated}
	service := newTestService(t, repository, &stubHasher{})
	handler := NewHandler(service, slog.New(slog.NewTextHandler(io.Discard, nil)), CookieConfig{Name: "aurora_session"})
	called := false
	protected := handler.Authenticate(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))
	for _, withCookie := range []bool{false, true} {
		request := httptest.NewRequest(http.MethodGet, "/api/cart", nil)
		if withCookie {
			request.AddCookie(&http.Cookie{Name: "aurora_session", Value: "raw-token"})
		}
		response := httptest.NewRecorder()
		protected.ServeHTTP(response, request)
		if response.Code != http.StatusUnauthorized || called {
			t.Fatalf("withCookie=%v: status=%d handlerCalled=%v", withCookie, response.Code, called)
		}
	}
}
