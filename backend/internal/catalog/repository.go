package catalog

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

type Repository interface {
	Create(context.Context, CreateProductInput) (Product, error)
	GetBySlug(context.Context, string) (Product, error)
	List(context.Context, int32, int32) (Page, error)
	GetSKUs(context.Context, []uuid.UUID) ([]SKUDetails, error)
}

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (r *PostgresRepository) Create(ctx context.Context, input CreateProductInput) (product Product, err error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return Product{}, fmt.Errorf("begin catalog transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	queries := db.New(tx)
	row, err := queries.CreateProduct(ctx, db.CreateProductParams{
		ID: uuid.New(), Name: input.Name, Slug: input.Slug, Description: input.Description,
	})
	if err != nil {
		return Product{}, mapWriteError(err)
	}

	product = productFromDB(row)
	product.SKUs = make([]SKU, 0, len(input.SKUs))
	for _, skuInput := range input.SKUs {
		skuRow, createErr := queries.CreateSKU(ctx, db.CreateSKUParams{
			ID: uuid.New(), ProductID: row.ID, Code: skuInput.Code,
			PriceMinor: skuInput.PriceMinor, Currency: skuInput.Currency,
		})
		if createErr != nil {
			return Product{}, mapWriteError(createErr)
		}
		product.SKUs = append(product.SKUs, skuFromDB(skuRow))
	}
	if err := tx.Commit(ctx); err != nil {
		return Product{}, fmt.Errorf("commit catalog transaction: %w", err)
	}
	return product, nil
}

func (r *PostgresRepository) List(ctx context.Context, page, pageSize int32) (Page, error) {
	queries := db.New(r.pool)
	total, err := queries.CountProducts(ctx)
	if err != nil {
		return Page{}, fmt.Errorf("count products: %w", err)
	}
	rows, err := queries.ListProducts(ctx, db.ListProductsParams{
		Limit: pageSize, Offset: (page - 1) * pageSize,
	})
	if err != nil {
		return Page{}, fmt.Errorf("list products: %w", err)
	}

	products := make([]Product, 0, len(rows))
	productIDs := make([]uuid.UUID, 0, len(rows))
	positions := make(map[uuid.UUID]int, len(rows))
	for _, row := range rows {
		positions[row.ID] = len(products)
		productIDs = append(productIDs, row.ID)
		products = append(products, productFromDB(row))
	}
	if len(productIDs) > 0 {
		skuRows, listErr := queries.ListSKUsByProductIDs(ctx, productIDs)
		if listErr != nil {
			return Page{}, fmt.Errorf("list product SKUs: %w", listErr)
		}
		for _, skuRow := range skuRows {
			position := positions[skuRow.ProductID]
			products[position].SKUs = append(products[position].SKUs, skuFromDB(skuRow))
		}
	}

	totalPages := int64(0)
	if total > 0 {
		totalPages = (total + int64(pageSize) - 1) / int64(pageSize)
	}
	return Page{Products: products, Page: page, PageSize: pageSize, TotalItems: total, TotalPages: totalPages}, nil
}

func (r *PostgresRepository) GetBySlug(ctx context.Context, slug string) (Product, error) {
	queries := db.New(r.pool)
	row, err := queries.GetProductBySlug(ctx, slug)
	if errors.Is(err, pgx.ErrNoRows) {
		return Product{}, ErrNotFound
	}
	if err != nil {
		return Product{}, fmt.Errorf("get product by slug: %w", err)
	}

	product := productFromDB(row)
	skuRows, err := queries.ListSKUsByProductIDs(ctx, []uuid.UUID{row.ID})
	if err != nil {
		return Product{}, fmt.Errorf("list product SKUs: %w", err)
	}
	for _, skuRow := range skuRows {
		product.SKUs = append(product.SKUs, skuFromDB(skuRow))
	}
	return product, nil
}

func (r *PostgresRepository) GetSKUs(ctx context.Context, ids []uuid.UUID) ([]SKUDetails, error) {
	if len(ids) == 0 {
		return []SKUDetails{}, nil
	}
	rows, err := db.New(r.pool).ListSKUDetailsByIDs(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("list catalog SKU details: %w", err)
	}
	details := make([]SKUDetails, 0, len(rows))
	for _, row := range rows {
		details = append(details, SKUDetails{
			ID: row.ID, ProductID: row.ProductID, ProductName: row.ProductName,
			ProductSlug: row.ProductSlug, Code: row.Code, PriceMinor: row.PriceMinor,
			Currency: row.Currency,
		})
	}
	return details, nil
}

func productFromDB(row db.Product) Product {
	return Product{
		ID: row.ID, Name: row.Name, Slug: row.Slug, Description: row.Description,
		SKUs: []SKU{}, CreatedAt: row.CreatedAt.Time, UpdatedAt: row.UpdatedAt.Time,
	}
}

func skuFromDB(row db.Sku) SKU {
	return SKU{
		ID: row.ID, ProductID: row.ProductID, Code: row.Code, PriceMinor: row.PriceMinor,
		Currency: row.Currency, CreatedAt: row.CreatedAt.Time, UpdatedAt: row.UpdatedAt.Time,
	}
}

func mapWriteError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return ErrConflict
	}
	return fmt.Errorf("write catalog data: %w", err)
}
