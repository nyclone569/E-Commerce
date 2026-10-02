-- +goose Up
ALTER TABLE skus RENAME COLUMN price_cents TO price_minor;

-- Existing non-VND development prices remain readable until explicitly repriced.
-- PostgreSQL enforces this constraint for all new or updated rows.
ALTER TABLE skus ADD CONSTRAINT skus_vnd_currency_check CHECK (currency = 'VND') NOT VALID;
-- Keep a maximum cart (100 SKUs x 99 each) below JavaScript's safe integer range.
ALTER TABLE skus ADD CONSTRAINT skus_price_minor_limit_check CHECK (price_minor <= 900000000000) NOT VALID;

-- +goose Down
ALTER TABLE skus DROP CONSTRAINT skus_vnd_currency_check;
ALTER TABLE skus DROP CONSTRAINT skus_price_minor_limit_check;
ALTER TABLE skus RENAME COLUMN price_minor TO price_cents;
