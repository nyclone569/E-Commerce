package identity

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/aurora-shop/aurora-shop/backend/internal/platform/httpx"
)

type userContextKey struct{}

type CookieConfig struct {
	Name   string
	Secure bool
}

type Handler struct {
	service *Service
	logger  *slog.Logger
	cookie  CookieConfig
}

func NewHandler(service *Service, logger *slog.Logger, cookie CookieConfig) *Handler {
	return &Handler{service: service, logger: logger, cookie: cookie}
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var input RegisterInput
	if err := httpx.DecodeJSON(r, &input); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid_json", "Request body must be a valid JSON object", nil)
		return
	}
	result, err := h.service.Register(r.Context(), input)
	if err != nil {
		h.writeServiceError(w, r, err)
		return
	}
	h.setSessionCookie(w, result.Token, result.ExpiresAt)
	httpx.WriteJSON(w, http.StatusCreated, map[string]User{"data": result.User})
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var input LoginInput
	if err := httpx.DecodeJSON(r, &input); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid_json", "Request body must be a valid JSON object", nil)
		return
	}
	result, err := h.service.Login(r.Context(), input)
	if err != nil {
		h.writeServiceError(w, r, err)
		return
	}
	h.setSessionCookie(w, result.Token, result.ExpiresAt)
	httpx.WriteJSON(w, http.StatusOK, map[string]User{"data": result.User})
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	user, ok := UserFromContext(r.Context())
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "authentication_required", "Authentication is required", nil)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]User{"data": user})
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(h.cookie.Name)
	if err == nil {
		if err := h.service.Logout(r.Context(), cookie.Value); err != nil {
			h.writeServiceError(w, r, err)
			return
		}
	}
	h.clearSessionCookie(w)
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(h.cookie.Name)
		if err != nil {
			httpx.WriteError(w, http.StatusUnauthorized, "authentication_required", "Authentication is required", nil)
			return
		}
		user, err := h.service.Authenticate(r.Context(), cookie.Value)
		if errors.Is(err, ErrUnauthenticated) {
			h.clearSessionCookie(w)
			httpx.WriteError(w, http.StatusUnauthorized, "authentication_required", "Authentication is required", nil)
			return
		}
		if err != nil {
			h.logger.Error("identity authentication failed", "request_id", httpx.RequestIDFromContext(r.Context()), "error", err)
			httpx.WriteError(w, http.StatusInternalServerError, "internal_error", "An internal error occurred", nil)
			return
		}
		ctx := context.WithValue(r.Context(), userContextKey{}, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func UserFromContext(ctx context.Context) (User, bool) {
	user, ok := ctx.Value(userContextKey{}).(User)
	return user, ok
}

func (h *Handler) writeServiceError(w http.ResponseWriter, r *http.Request, err error) {
	var validationErr *ValidationError
	switch {
	case errors.As(err, &validationErr):
		httpx.WriteError(w, http.StatusBadRequest, "validation_failed", "Request validation failed", validationErr.Fields)
	case errors.Is(err, ErrConflict):
		httpx.WriteError(w, http.StatusConflict, "identity_conflict", "An account with this email already exists", nil)
	case errors.Is(err, ErrInvalidCredentials):
		httpx.WriteError(w, http.StatusUnauthorized, "invalid_credentials", "Invalid email or password", nil)
	default:
		h.logger.Error("identity request failed", "request_id", httpx.RequestIDFromContext(r.Context()), "error", err)
		httpx.WriteError(w, http.StatusInternalServerError, "internal_error", "An internal error occurred", nil)
	}
}

func (h *Handler) setSessionCookie(w http.ResponseWriter, token string, expiresAt time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name: h.cookie.Name, Value: token, Path: "/", Expires: expiresAt,
		MaxAge: int(time.Until(expiresAt).Seconds()), HttpOnly: true, Secure: h.cookie.Secure,
		SameSite: http.SameSiteLaxMode,
	})
}

func (h *Handler) clearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name: h.cookie.Name, Value: "", Path: "/", Expires: time.Unix(1, 0),
		MaxAge: -1, HttpOnly: true, Secure: h.cookie.Secure, SameSite: http.SameSiteLaxMode,
	})
}
