//go:build integration

package integration_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/aurora-shop/aurora-shop/backend/internal/cart"
	"github.com/aurora-shop/aurora-shop/backend/internal/catalog"
	"github.com/google/uuid"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestCartOwnershipAndConcurrentAdd(t *testing.T) {
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

	ownerID, otherID := uuid.New(), uuid.New()
	for _, user := range []struct {
		id    uuid.UUID
		email string
	}{{ownerID, "owner@example.com"}, {otherID, "other@example.com"}} {
		if _, err := pool.Exec(ctx, `INSERT INTO users (id, email, normalized_email, password_hash, display_name) VALUES ($1, $2, $2, 'test-only', 'Test')`, user.id, user.email); err != nil {
			t.Fatalf("insert test user: %v", err)
		}
	}
	catalogService := catalog.NewService(catalog.NewPostgresRepository(pool))
	product := createProduct(t, ctx, catalogService, "Cart Mug", "cart-mug", "CART-MUG")
	skuID := product.SKUs[0].ID
	repository := cart.NewPostgresRepository(pool)
	service := cart.NewService(repository, catalogService)

	empty, err := service.Get(ctx, ownerID)
	if err != nil || len(empty.Items) != 0 || empty.Currency != "VND" {
		t.Fatalf("empty cart = %#v, %v", empty, err)
	}
	var cartCount int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM carts`).Scan(&cartCount); err != nil || cartCount != 0 {
		t.Fatalf("GET created a cart row: count=%d err=%v", cartCount, err)
	}
	if err := service.Add(ctx, ownerID, skuID, 1); err != nil {
		t.Fatalf("first add: %v", err)
	}
	if err := service.Add(ctx, ownerID, skuID, 2); err != nil {
		t.Fatalf("second add: %v", err)
	}

	other, err := service.Get(ctx, otherID)
	if err != nil || len(other.Items) != 0 {
		t.Fatalf("other user's cart = %#v, %v", other, err)
	}
	if err := service.SetQuantity(ctx, otherID, skuID, 5); !errors.Is(err, cart.ErrItemNotFound) {
		t.Fatalf("cross-user update error = %v", err)
	}
	if err := service.Remove(ctx, otherID, skuID); err != nil {
		t.Fatalf("cross-user remove: %v", err)
	}

	const concurrentAdds = 20
	var wg sync.WaitGroup
	failures := make(chan error, concurrentAdds)
	for i := 0; i < concurrentAdds; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			failures <- service.Add(ctx, ownerID, skuID, 1)
		}()
	}
	wg.Wait()
	close(failures)
	for err := range failures {
		if err != nil {
			t.Fatalf("concurrent add: %v", err)
		}
	}
	current, err := service.Get(ctx, ownerID)
	if err != nil || len(current.Items) != 1 || current.Items[0].Quantity != 23 || current.SubtotalMinor != 23*249000 {
		t.Fatalf("cart after concurrent add = %#v, %v", current, err)
	}
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM carts WHERE user_id=$1 AND status='active'`, ownerID).Scan(&cartCount); err != nil || cartCount != 1 {
		t.Fatalf("active cart count=%d err=%v", cartCount, err)
	}
	if err := service.SetQuantity(ctx, ownerID, skuID, 99); err != nil {
		t.Fatalf("set quantity: %v", err)
	}
	if err := service.Add(ctx, ownerID, skuID, 1); !errors.Is(err, cart.ErrQuantityLimit) {
		t.Fatalf("quantity overflow error = %v", err)
	}
	if err := service.Remove(ctx, ownerID, skuID); err != nil {
		t.Fatalf("remove: %v", err)
	}
	empty, err = service.Get(ctx, ownerID)
	if err != nil || len(empty.Items) != 0 {
		t.Fatalf("cart after remove = %#v, %v", empty, err)
	}
}
