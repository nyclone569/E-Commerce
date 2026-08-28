package httpx

import (
	"context"
	"log/slog"
	"net/http"
	"net/url"
	"runtime/debug"
	"time"

	"github.com/google/uuid"
)

func RequireOrigin(allowedOrigin string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin == "" || !sameOrigin(origin, allowedOrigin) {
				WriteError(w, http.StatusForbidden, "origin_not_allowed", "Request origin is not allowed", nil)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func sameOrigin(candidate, allowed string) bool {
	candidateURL, candidateErr := url.Parse(candidate)
	allowedURL, allowedErr := url.Parse(allowed)
	if candidateErr != nil || allowedErr != nil {
		return false
	}
	return candidateURL.Scheme == allowedURL.Scheme && candidateURL.Host == allowedURL.Host && candidateURL.Path == "" && candidateURL.RawQuery == "" && candidateURL.Fragment == ""
}

type contextKey string

const requestIDKey contextKey = "request_id"

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" || len(requestID) > 128 {
			requestID = uuid.NewString()
		}
		w.Header().Set("X-Request-ID", requestID)
		ctx := context.WithValue(r.Context(), requestIDKey, requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func RequestIDFromContext(ctx context.Context) string {
	value, _ := ctx.Value(requestIDKey).(string)
	return value
}

func Recoverer(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if recovered := recover(); recovered != nil {
					logger.Error("panic recovered", "request_id", RequestIDFromContext(r.Context()), "panic", recovered, "stack", string(debug.Stack()))
					WriteError(w, http.StatusInternalServerError, "internal_error", "An internal error occurred", nil)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

func AccessLog(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			started := time.Now()
			recorder := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(recorder, r)
			logger.Info("request completed",
				"request_id", RequestIDFromContext(r.Context()),
				"method", r.Method,
				"route", r.URL.Path,
				"status_code", recorder.status,
				"duration_ms", time.Since(started).Milliseconds(),
			)
		})
	}
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}
