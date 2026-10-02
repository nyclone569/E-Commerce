package cart

import "github.com/google/uuid"

type Item struct {
	SKUID          uuid.UUID `json:"sku_id"`
	ProductID      uuid.UUID `json:"product_id"`
	ProductName    string    `json:"product_name"`
	ProductSlug    string    `json:"product_slug"`
	SKUCode        string    `json:"sku_code"`
	Quantity       int32     `json:"quantity"`
	UnitPriceMinor int64     `json:"unit_price_minor"`
	LineTotalMinor int64     `json:"line_total_minor"`
}

type Cart struct {
	Items         []Item `json:"items"`
	SubtotalMinor int64  `json:"subtotal_minor"`
	Currency      string `json:"currency"`
}

type StoredItem struct {
	SKUID    uuid.UUID
	Quantity int32
}
