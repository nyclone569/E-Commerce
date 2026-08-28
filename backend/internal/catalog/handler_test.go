package catalog

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestCreateProductReturnsConsistentValidationError(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	handler := NewHandler(NewService(&stubRepository{}), logger)
	request := httptest.NewRequest(http.MethodPost, "/api/products", bytes.NewBufferString(`{"name":"","slug":"bad slug","skus":[]}`))
	response := httptest.NewRecorder()

	handler.CreateProduct(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
	}
	if body := response.Body.String(); !strings.Contains(body, `"code":"validation_failed"`) || !strings.Contains(body, `"details"`) {
		t.Fatalf("unexpected response: %s", body)
	}
}

func TestCreateProductRejectsUnknownJSONFields(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	handler := NewHandler(NewService(&stubRepository{}), logger)
	request := httptest.NewRequest(http.MethodPost, "/api/products", bytes.NewBufferString(`{"name":"Mug","slug":"mug","unexpected":true,"skus":[]}`))
	response := httptest.NewRecorder()

	handler.CreateProduct(response, request)

	if response.Code != http.StatusBadRequest || !strings.Contains(response.Body.String(), `"code":"invalid_json"`) {
		t.Fatalf("status/body = %d %s", response.Code, response.Body.String())
	}
}

func TestGetProductReturnsNotFoundEnvelope(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	handler := NewHandler(NewService(&stubRepository{getErr: ErrNotFound}), logger)
	router := chi.NewRouter()
	router.Get("/api/products/{slug}", handler.GetProduct)
	request := httptest.NewRequest(http.MethodGet, "/api/products/missing-product", nil)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusNotFound || !strings.Contains(response.Body.String(), `"code":"product_not_found"`) {
		t.Fatalf("status/body = %d %s", response.Code, response.Body.String())
	}
}
