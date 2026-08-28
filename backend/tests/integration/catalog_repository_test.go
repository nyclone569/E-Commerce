//go:build integration

package integration_test

import (
	"context"
	"database/sql"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/aurora-shop/aurora-shop/backend/internal/catalog"
	platformdatabase "github.com/aurora-shop/aurora-shop/backend/internal/platform/database"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestCatalogRepositoryCreateAndList(t *testing.T) {
	testcontainers.SkipIfProviderIsNotHealthy(t)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	container, err := postgres.Run(ctx, "postgres:17-alpine",
		postgres.WithDatabase("aurora"), postgres.WithUsername("aurora"), postgres.WithPassword("aurora"),
		testcontainers.WithWaitStrategy(wait.ForLog("database system is ready to accept connections").WithOccurrence(2).WithStartupTimeout(60*time.Second)),
	)
	if err != nil {
		t.Fatalf("start PostgreSQL container: %v", err)
	}
	t.Cleanup(func() { _ = testcontainers.TerminateContainer(container) })

	databaseURL, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("database connection string: %v", err)
	}
	applyMigrations(t, databaseURL)

	pool := openPool(t, ctx, databaseURL)
	repository := catalog.NewPostgresRepository(pool)
	service := catalog.NewService(repository)
	first := createProduct(t, ctx, service, "Aurora Mug", "aurora-mug", "MUG-WHITE")
	second := createProduct(t, ctx, service, "Aurora Tee", "aurora-tee", "TEE-BLACK-M")

	page, err := service.ListProducts(ctx, 1, 1)
	if err != nil {
		t.Fatalf("ListProducts() error = %v", err)
	}
	if page.TotalItems != 2 || page.TotalPages != 2 || len(page.Products) != 1 {
		t.Fatalf("unexpected page: %#v", page)
	}
	if page.Products[0].ID != second.ID || page.Products[0].ID == first.ID {
		t.Fatalf("products are not deterministically newest-first: %#v", page.Products)
	}
	if len(page.Products[0].SKUs) != 1 || page.Products[0].SKUs[0].Code != "TEE-BLACK-M" {
		t.Fatalf("SKUs were not loaded: %#v", page.Products[0].SKUs)
	}

	detail, err := service.GetProduct(ctx, "aurora-mug")
	if err != nil {
		t.Fatalf("GetProduct() error = %v", err)
	}
	if detail.ID != first.ID || len(detail.SKUs) != 1 || detail.SKUs[0].Code != "MUG-WHITE" {
		t.Fatalf("unexpected product detail: %#v", detail)
	}
	_, err = service.GetProduct(ctx, "missing-product")
	if err != catalog.ErrNotFound {
		t.Fatalf("missing GetProduct() error = %v, want ErrNotFound", err)
	}

	_, err = service.CreateProduct(ctx, catalog.CreateProductInput{
		Name: "Duplicate", Slug: "aurora-mug",
		SKUs: []catalog.CreateSKUInput{{Code: "ANOTHER", PriceCents: 100, Currency: "USD"}},
	})
	if err != catalog.ErrConflict {
		t.Fatalf("duplicate CreateProduct() error = %v, want ErrConflict", err)
	}

	_, err = service.CreateProduct(ctx, catalog.CreateProductInput{
		Name: "Rollback Candidate", Slug: "rollback-candidate",
		SKUs: []catalog.CreateSKUInput{{Code: "MUG-WHITE", PriceCents: 100, Currency: "USD"}},
	})
	if err != catalog.ErrConflict {
		t.Fatalf("duplicate SKU CreateProduct() error = %v, want ErrConflict", err)
	}
	allProducts, err := service.ListProducts(ctx, 1, 10)
	if err != nil {
		t.Fatalf("ListProducts() after rollback error = %v", err)
	}
	if allProducts.TotalItems != 2 {
		t.Fatalf("failed SKU insert left a partial product: total = %d, want 2", allProducts.TotalItems)
	}
}

func createProduct(t *testing.T, ctx context.Context, service *catalog.Service, name, slug, code string) catalog.Product {
	t.Helper()
	product, err := service.CreateProduct(ctx, catalog.CreateProductInput{
		Name: name, Slug: slug, Description: "Milestone 1 product",
		SKUs: []catalog.CreateSKUInput{{Code: code, PriceCents: 2499, Currency: "USD"}},
	})
	if err != nil {
		t.Fatalf("CreateProduct() error = %v", err)
	}
	time.Sleep(2 * time.Millisecond)
	return product
}

func applyMigrations(t *testing.T, databaseURL string) {
	t.Helper()
	database, err := sql.Open("pgx", databaseURL)
	if err != nil {
		t.Fatalf("open migration database: %v", err)
	}
	defer database.Close()
	_, filename, _, _ := runtime.Caller(0)
	migrationDir := filepath.Join(filepath.Dir(filename), "..", "..", "db", "migrations")
	if err := goose.Up(database, migrationDir); err != nil {
		t.Fatalf("apply migrations: %v", err)
	}
}

func openPool(t *testing.T, ctx context.Context, databaseURL string) *pgxpool.Pool {
	t.Helper()
	pool, err := platformdatabase.Open(ctx, databaseURL, 5, 1, 10*time.Second)
	if err != nil {
		t.Fatalf("open application pool: %v", err)
	}
	t.Cleanup(pool.Close)
	return pool
}
