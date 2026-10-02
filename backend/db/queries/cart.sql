-- name: GetActiveCartItemsByUser :many
SELECT ci.sku_id, ci.quantity
FROM carts AS c
JOIN cart_items AS ci ON ci.cart_id = c.id
WHERE c.user_id = $1 AND c.status = 'active'
ORDER BY ci.created_at, ci.sku_id;

-- name: CreateOrGetActiveCart :one
INSERT INTO carts (id, user_id)
VALUES ($1, $2)
ON CONFLICT (user_id) WHERE status = 'active'
DO UPDATE SET updated_at = now()
RETURNING id;

-- name: AddCartItem :one
INSERT INTO cart_items (cart_id, sku_id, quantity)
VALUES ($1, $2, $3)
ON CONFLICT (cart_id, sku_id)
DO UPDATE SET quantity = cart_items.quantity + EXCLUDED.quantity,
              updated_at = now()
WHERE cart_items.quantity + EXCLUDED.quantity <= 99
RETURNING quantity;

-- name: CountCartItems :one
SELECT count(*) FROM cart_items WHERE cart_id = $1;

-- name: SetCartItemQuantity :one
UPDATE cart_items
SET quantity = $3, updated_at = now()
WHERE sku_id = $2
  AND cart_id = (SELECT id FROM carts WHERE user_id = $1 AND status = 'active')
RETURNING quantity;

-- name: RemoveCartItem :exec
DELETE FROM cart_items
WHERE sku_id = $2
  AND cart_id = (SELECT id FROM carts WHERE user_id = $1 AND status = 'active');
