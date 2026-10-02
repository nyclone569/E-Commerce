package cart

import (
	"context"
	"errors"
	"fmt"

	"github.com/aurora-shop/aurora-shop/backend/internal/db"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (r *PostgresRepository) List(ctx context.Context, userID uuid.UUID) ([]StoredItem, error) {
	rows, err := db.New(r.pool).GetActiveCartItemsByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list cart items: %w", err)
	}
	items := make([]StoredItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, StoredItem{SKUID: row.SkuID, Quantity: row.Quantity})
	}
	return items, nil
}

func (r *PostgresRepository) Add(ctx context.Context, userID, skuID uuid.UUID, quantity int32) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin cart transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	queries := db.New(tx)
	cartID, err := queries.CreateOrGetActiveCart(ctx, db.CreateOrGetActiveCartParams{ID: uuid.New(), UserID: userID})
	if err != nil {
		return fmt.Errorf("create active cart: %w", err)
	}
	if _, err := queries.AddCartItem(ctx, db.AddCartItemParams{CartID: cartID, SkuID: skuID, Quantity: quantity}); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrQuantityLimit
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" && pgErr.ConstraintName == "cart_items_sku_id_fkey" {
			return ErrSKUUnavailable
		}
		return fmt.Errorf("add cart item: %w", err)
	}
	count, err := queries.CountCartItems(ctx, cartID)
	if err != nil {
		return fmt.Errorf("count cart items: %w", err)
	}
	if count > 100 {
		return ErrCartLimit
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit cart transaction: %w", err)
	}
	return nil
}

func (r *PostgresRepository) SetQuantity(ctx context.Context, userID, skuID uuid.UUID, quantity int32) error {
	_, err := db.New(r.pool).SetCartItemQuantity(ctx, db.SetCartItemQuantityParams{
		UserID: userID, SkuID: skuID, Quantity: quantity,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrItemNotFound
	}
	if err != nil {
		return fmt.Errorf("set cart quantity: %w", err)
	}
	return nil
}

func (r *PostgresRepository) Remove(ctx context.Context, userID, skuID uuid.UUID) error {
	if err := db.New(r.pool).RemoveCartItem(ctx, db.RemoveCartItemParams{UserID: userID, SkuID: skuID}); err != nil {
		return fmt.Errorf("remove cart item: %w", err)
	}
	return nil
}
