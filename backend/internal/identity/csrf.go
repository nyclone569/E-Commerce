package identity

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"net/http"

	"github.com/aurora-shop/aurora-shop/backend/internal/platform/httpx"
)

const csrfPurpose = "aurora-shop/csrf/v1"

func csrfToken(rawSession string) string {
	mac := hmac.New(sha256.New, []byte(rawSession))
	_, _ = mac.Write([]byte(csrfPurpose))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func (h *Handler) CSRFToken(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(h.cookie.Name)
	if err != nil {
		httpx.WriteError(w, http.StatusUnauthorized, "authentication_required", "Authentication is required", nil)
		return
	}
	w.Header().Set("Cache-Control", "private, no-store")
	w.Header().Set("Vary", "Cookie")
	httpx.WriteJSON(w, http.StatusOK, map[string]string{"csrf_token": csrfToken(cookie.Value)})
}

func (h *Handler) RequireCSRF(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(h.cookie.Name)
		provided := r.Header.Get("X-CSRF-Token")
		if err != nil || provided == "" {
			httpx.WriteError(w, http.StatusForbidden, "csrf_invalid", "CSRF token is invalid", nil)
			return
		}
		expected := csrfToken(cookie.Value)
		if !hmac.Equal([]byte(provided), []byte(expected)) {
			httpx.WriteError(w, http.StatusForbidden, "csrf_invalid", "CSRF token is invalid", nil)
			return
		}
		next.ServeHTTP(w, r)
	})
}
