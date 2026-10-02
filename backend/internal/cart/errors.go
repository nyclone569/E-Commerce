package cart

import "errors"

var (
	ErrSKUUnavailable = errors.New("SKU is unavailable for cart")
	ErrItemNotFound   = errors.New("cart item was not found")
	ErrQuantityLimit  = errors.New("cart quantity limit exceeded")
	ErrCartLimit      = errors.New("cart item limit exceeded")
)

type ValidationError struct {
	Fields map[string]string
}

func (e *ValidationError) Error() string { return "cart validation failed" }
