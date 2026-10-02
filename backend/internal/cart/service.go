package cart

import (
	"context"
	"fmt"
	"math"

	"github.com/aurora-shop/aurora-shop/backend/internal/catalog"
	"github.com/google/uuid"
)

const maxQuantity = 99

type Repository interface {
	List(context.Context, uuid.UUID) ([]StoredItem, error)
	Add(context.Context, uuid.UUID, uuid.UUID, int32) error
	SetQuantity(context.Context, uuid.UUID, uuid.UUID, int32) error
	Remove(context.Context, uuid.UUID, uuid.UUID) error
}

type CatalogLookup interface {
	GetSKUs(context.Context, []uuid.UUID) ([]catalog.SKUDetails, error)
}

type Service struct {
	repository Repository
	catalog    CatalogLookup
}

func NewService(repository Repository, catalog CatalogLookup) *Service {
	return &Service{repository: repository, catalog: catalog}
}

func (s *Service) Get(ctx context.Context, userID uuid.UUID) (Cart, error) {
	stored, err := s.repository.List(ctx, userID)
	if err != nil {
		return Cart{}, err
	}
	result := Cart{Items: make([]Item, 0, len(stored)), Currency: "VND"}
	if len(stored) == 0 {
		return result, nil
	}
	ids := make([]uuid.UUID, 0, len(stored))
	for _, item := range stored {
		ids = append(ids, item.SKUID)
	}
	skus, err := s.catalog.GetSKUs(ctx, ids)
	if err != nil {
		return Cart{}, fmt.Errorf("load cart SKU details: %w", err)
	}
	byID := make(map[uuid.UUID]catalog.SKUDetails, len(skus))
	for _, sku := range skus {
		byID[sku.ID] = sku
	}
	for _, storedItem := range stored {
		sku, found := byID[storedItem.SKUID]
		if !found || sku.Currency != "VND" {
			return Cart{}, ErrSKUUnavailable
		}
		if sku.PriceMinor > math.MaxInt64/int64(storedItem.Quantity) {
			return Cart{}, fmt.Errorf("cart line total overflows int64")
		}
		lineTotal := sku.PriceMinor * int64(storedItem.Quantity)
		if result.SubtotalMinor > math.MaxInt64-lineTotal {
			return Cart{}, fmt.Errorf("cart subtotal overflows int64")
		}
		result.SubtotalMinor += lineTotal
		result.Items = append(result.Items, Item{
			SKUID: sku.ID, ProductID: sku.ProductID, ProductName: sku.ProductName,
			ProductSlug: sku.ProductSlug, SKUCode: sku.Code, Quantity: storedItem.Quantity,
			UnitPriceMinor: sku.PriceMinor, LineTotalMinor: lineTotal,
		})
	}
	return result, nil
}

func (s *Service) Add(ctx context.Context, userID, skuID uuid.UUID, quantity int32) error {
	if err := validate(userID, skuID, quantity); err != nil {
		return err
	}
	skus, err := s.catalog.GetSKUs(ctx, []uuid.UUID{skuID})
	if err != nil {
		return fmt.Errorf("validate cart SKU: %w", err)
	}
	if len(skus) != 1 || skus[0].Currency != "VND" {
		return ErrSKUUnavailable
	}
	return s.repository.Add(ctx, userID, skuID, quantity)
}

func (s *Service) SetQuantity(ctx context.Context, userID, skuID uuid.UUID, quantity int32) error {
	if err := validate(userID, skuID, quantity); err != nil {
		return err
	}
	return s.repository.SetQuantity(ctx, userID, skuID, quantity)
}

func (s *Service) Remove(ctx context.Context, userID, skuID uuid.UUID) error {
	if userID == uuid.Nil || skuID == uuid.Nil {
		return &ValidationError{Fields: map[string]string{"sku_id": "must be a valid UUID"}}
	}
	return s.repository.Remove(ctx, userID, skuID)
}

func validate(userID, skuID uuid.UUID, quantity int32) error {
	fields := map[string]string{}
	if userID == uuid.Nil || skuID == uuid.Nil {
		fields["sku_id"] = "must be a valid UUID"
	}
	if quantity < 1 || quantity > maxQuantity {
		fields["quantity"] = "must be between 1 and 99"
	}
	if len(fields) > 0 {
		return &ValidationError{Fields: fields}
	}
	return nil
}
