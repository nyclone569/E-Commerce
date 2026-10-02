package cart

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/aurora-shop/aurora-shop/backend/internal/identity"
	"github.com/aurora-shop/aurora-shop/backend/internal/platform/httpx"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type Handler struct {
	service *Service
	logger  *slog.Logger
}

func NewHandler(service *Service, logger *slog.Logger) *Handler {
	return &Handler{service: service, logger: logger}
}

type quantityInput struct {
	Quantity int32 `json:"quantity"`
}

type addItemInput struct {
	SKUID    uuid.UUID `json:"sku_id"`
	Quantity int32     `json:"quantity"`
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(r)
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "authentication_required", "Authentication is required", nil)
		return
	}
	w.Header().Set("Cache-Control", "private, no-store")
	result, err := h.service.Get(r.Context(), userID)
	if err != nil {
		h.writeError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, result)
}

func (h *Handler) Add(w http.ResponseWriter, r *http.Request) {
	userID, ok := currentUserID(r)
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "authentication_required", "Authentication is required", nil)
		return
	}
	var input addItemInput
	if err := httpx.DecodeJSON(r, &input); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid_json", "Request body must be a valid JSON object", nil)
		return
	}
	if err := h.service.Add(r.Context(), userID, input.SKUID, input.Quantity); err != nil {
		h.writeError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) SetQuantity(w http.ResponseWriter, r *http.Request) {
	userID, skuID, ok := parseItemIdentity(w, r)
	if !ok {
		return
	}
	var input quantityInput
	if err := httpx.DecodeJSON(r, &input); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "invalid_json", "Request body must be a valid JSON object", nil)
		return
	}
	if err := h.service.SetQuantity(r.Context(), userID, skuID, input.Quantity); err != nil {
		h.writeError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) Remove(w http.ResponseWriter, r *http.Request) {
	userID, skuID, ok := parseItemIdentity(w, r)
	if !ok {
		return
	}
	if err := h.service.Remove(r.Context(), userID, skuID); err != nil {
		h.writeError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func currentUserID(r *http.Request) (uuid.UUID, bool) {
	user, ok := identity.UserFromContext(r.Context())
	return user.ID, ok
}

func parseItemIdentity(w http.ResponseWriter, r *http.Request) (uuid.UUID, uuid.UUID, bool) {
	userID, ok := currentUserID(r)
	if !ok {
		httpx.WriteError(w, http.StatusUnauthorized, "authentication_required", "Authentication is required", nil)
		return uuid.Nil, uuid.Nil, false
	}
	skuID, err := uuid.Parse(chi.URLParam(r, "skuID"))
	if err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "validation_failed", "Request validation failed", map[string]string{"sku_id": "must be a valid UUID"})
		return uuid.Nil, uuid.Nil, false
	}
	return userID, skuID, true
}

func (h *Handler) writeError(w http.ResponseWriter, r *http.Request, err error) {
	var validationErr *ValidationError
	switch {
	case errors.As(err, &validationErr):
		httpx.WriteError(w, http.StatusBadRequest, "validation_failed", "Request validation failed", validationErr.Fields)
	case errors.Is(err, ErrSKUUnavailable):
		httpx.WriteError(w, http.StatusConflict, "sku_unavailable", "SKU is unavailable for cart", nil)
	case errors.Is(err, ErrItemNotFound):
		httpx.WriteError(w, http.StatusNotFound, "cart_item_not_found", "Cart item was not found", nil)
	case errors.Is(err, ErrQuantityLimit), errors.Is(err, ErrCartLimit):
		httpx.WriteError(w, http.StatusConflict, "cart_limit", "Cart limit was exceeded", nil)
	default:
		h.logger.Error("cart request failed", "request_id", httpx.RequestIDFromContext(r.Context()), "error", err)
		httpx.WriteError(w, http.StatusInternalServerError, "internal_error", "An internal error occurred", nil)
	}
}
