package cart

import (
	"context"
	"errors"
	"testing"

	"github.com/aurora-shop/aurora-shop/backend/internal/catalog"
	"github.com/google/uuid"
)

type fakeRepository struct {
	items []StoredItem
	adds  int
}

func (r *fakeRepository) List(context.Context, uuid.UUID) ([]StoredItem, error) { return r.items, nil }
func (r *fakeRepository) Add(context.Context, uuid.UUID, uuid.UUID, int32) error {
	r.adds++
	return nil
}
func (r *fakeRepository) SetQuantity(context.Context, uuid.UUID, uuid.UUID, int32) error { return nil }
func (r *fakeRepository) Remove(context.Context, uuid.UUID, uuid.UUID) error             { return nil }

type fakeCatalog struct{ skus []catalog.SKUDetails }

func (c fakeCatalog) GetSKUs(context.Context, []uuid.UUID) ([]catalog.SKUDetails, error) {
	return c.skus, nil
}

func TestGetEmptyCartDoesNotCreateItems(t *testing.T) {
	result, err := NewService(&fakeRepository{}, fakeCatalog{}).Get(context.Background(), uuid.New())
	if err != nil || result.Items == nil || len(result.Items) != 0 || result.SubtotalMinor != 0 || result.Currency != "VND" {
		t.Fatalf("Get() = %#v, %v", result, err)
	}
}

func TestAddValidatesQuantityAndCurrency(t *testing.T) {
	userID, skuID := uuid.New(), uuid.New()
	repository := &fakeRepository{}
	service := NewService(repository, fakeCatalog{skus: []catalog.SKUDetails{{ID: skuID, Currency: "USD"}}})
	if err := service.Add(context.Background(), userID, skuID, 0); err == nil {
		t.Fatal("zero quantity accepted")
	}
	if err := service.Add(context.Background(), userID, skuID, 100); err == nil {
		t.Fatal("quantity above limit accepted")
	}
	if err := service.Add(context.Background(), userID, skuID, 1); !errors.Is(err, ErrSKUUnavailable) {
		t.Fatalf("USD SKU error = %v", err)
	}
	if repository.adds != 0 {
		t.Fatalf("repository was called %d times", repository.adds)
	}
}

func TestGetComputesCurrentVNDPrice(t *testing.T) {
	userID, skuID := uuid.New(), uuid.New()
	service := NewService(&fakeRepository{items: []StoredItem{{SKUID: skuID, Quantity: 2}}}, fakeCatalog{skus: []catalog.SKUDetails{{ID: skuID, Currency: "VND", PriceMinor: 249000}}})
	result, err := service.Get(context.Background(), userID)
	if err != nil || result.SubtotalMinor != 498000 || result.Items[0].LineTotalMinor != 498000 {
		t.Fatalf("Get() = %#v, %v", result, err)
	}
}
