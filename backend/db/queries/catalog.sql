-- name: CreateProduct :one
INSERT INTO products (id, name, slug, description)
VALUES ($1, $2, $3, $4)
RETURNING id, name, slug, description, created_at, updated_at;

-- name: CreateSKU :one
INSERT INTO skus (id, product_id, code, price_minor, currency)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, product_id, code, price_minor, currency, created_at, updated_at;

-- name: CountProducts :one
SELECT count(*) FROM products;

-- name: GetProductBySlug :one
SELECT id, name, slug, description, created_at, updated_at
FROM products
WHERE slug = $1;

-- name: ListProducts :many
SELECT id, name, slug, description, created_at, updated_at
FROM products
ORDER BY created_at DESC, id DESC
LIMIT $1 OFFSET $2;

-- name: ListSKUsByProductIDs :many
SELECT id, product_id, code, price_minor, currency, created_at, updated_at
FROM skus
WHERE product_id = ANY($1::uuid[])
ORDER BY product_id, code ASC;

-- name: ListSKUDetailsByIDs :many
SELECT s.id, s.product_id, p.name AS product_name, p.slug AS product_slug,
       s.code, s.price_minor, s.currency
FROM skus AS s
JOIN products AS p ON p.id = s.product_id
WHERE s.id = ANY($1::uuid[]);
