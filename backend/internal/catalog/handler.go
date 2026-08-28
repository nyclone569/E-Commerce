package catalog

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/aurora-shop/aurora-shop/backend/internal/platform/httpx"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	service *Service
	logger  *slog.Logger
}

func NewHandler(service *Service, logger *slog.Logger) *Handler {
	return &Handler{service: service, logger: logger}
}

func (h *Handler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	var input CreateProductInput
	if err := httpx.DecodeJSON(r, &input); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid_json", "Request body must be a valid JSON object", nil)
		return
	}
	product, err := h.service.CreateProduct(r.Context(), input)
	if err != nil {
		h.writeServiceError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, product)
}

func (h *Handler) ListProducts(w http.ResponseWriter, r *http.Request) {
	page, err := parseInt32Query(r, "page", 1)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "validation_failed", "Request validation failed", map[string]string{"page": "must be an integer"})
		return
	}
	pageSize, err := parseInt32Query(r, "page_size", 20)
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "validation_failed", "Request validation failed", map[string]string{"page_size": "must be an integer"})
		return
	}
	result, err := h.service.ListProducts(r.Context(), page, pageSize)
	if err != nil {
		h.writeServiceError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result)
}

func (h *Handler) GetProduct(w http.ResponseWriter, r *http.Request) {
	product, err := h.service.GetProduct(r.Context(), chi.URLParam(r, "slug"))
	if err != nil {
		h.writeServiceError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, product)
}

func (h *Handler) writeServiceError(w http.ResponseWriter, r *http.Request, err error) {
	var validationErr *ValidationError
	switch {
	case errors.As(err, &validationErr):
		httpx.WriteError(w, http.StatusBadRequest, "validation_failed", "Request validation failed", validationErr.Fields)
	case errors.Is(err, ErrConflict):
		httpx.WriteError(w, http.StatusConflict, "catalog_conflict", "A product slug or SKU code already exists", nil)
	case errors.Is(err, ErrNotFound):
		httpx.WriteError(w, http.StatusNotFound, "product_not_found", "Product was not found", nil)
	default:
		h.logger.Error("catalog request failed", "request_id", httpx.RequestIDFromContext(r.Context()), "error", err)
		httpx.WriteError(w, http.StatusInternalServerError, "internal_error", "An internal error occurred", nil)
	}
}

func parseInt32Query(r *http.Request, key string, fallback int32) (int32, error) {
	raw := r.URL.Query().Get(key)
	if raw == "" {
		return fallback, nil
	}
	value, err := strconv.ParseInt(raw, 10, 32)
	return int32(value), err
}
