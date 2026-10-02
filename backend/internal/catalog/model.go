package catalog

import (
	"time"

	"github.com/google/uuid"
)

type Product struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	Description string    `json:"description"`
	SKUs        []SKU     `json:"skus"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type SKU struct {
	ID         uuid.UUID `json:"id"`
	ProductID  uuid.UUID `json:"product_id"`
	Code       string    `json:"code"`
	PriceMinor int64     `json:"price_minor"`
	Currency   string    `json:"currency"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type CreateProductInput struct {
	Name        string           `json:"name"`
	Slug        string           `json:"slug"`
	Description string           `json:"description"`
	SKUs        []CreateSKUInput `json:"skus"`
}

type CreateSKUInput struct {
	Code       string `json:"code"`
	PriceMinor int64  `json:"price_minor"`
	Currency   string `json:"currency"`
}

type SKUDetails struct {
	ID          uuid.UUID
	ProductID   uuid.UUID
	ProductName string
	ProductSlug string
	Code        string
	PriceMinor  int64
	Currency    string
}

type Page struct {
	Products   []Product `json:"products"`
	Page       int32     `json:"page"`
	PageSize   int32     `json:"page_size"`
	TotalItems int64     `json:"total_items"`
	TotalPages int64     `json:"total_pages"`
}
