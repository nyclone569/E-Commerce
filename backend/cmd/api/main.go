package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aurora-shop/aurora-shop/backend/internal/cart"
	"github.com/aurora-shop/aurora-shop/backend/internal/catalog"
	"github.com/aurora-shop/aurora-shop/backend/internal/identity"
	"github.com/aurora-shop/aurora-shop/backend/internal/platform/config"
	"github.com/aurora-shop/aurora-shop/backend/internal/platform/database"
	"github.com/aurora-shop/aurora-shop/backend/internal/platform/httpx"
	"github.com/aurora-shop/aurora-shop/backend/internal/platform/logging"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	readTimeout     = 10 * time.Second
	readHeaderLimit = 5 * time.Second
	writeTimeout    = 15 * time.Second
	idleTimeout     = 60 * time.Second
	shutdownTimeout = 10 * time.Second
)

func main() {
	if err := run(); err != nil {
		slog.Error("API stopped", "error", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	logger := logging.New("aurora-shop-api", cfg.Environment, cfg.Version)
	slog.SetDefault(logger)

	rootCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	pool, err := database.Open(rootCtx, cfg.DatabaseURL, cfg.DatabaseMaxOpen, cfg.DatabaseMinIdle, cfg.DatabaseTimeout)
	if err != nil {
		return err
	}
	defer pool.Close()

	router, err := routes(pool, logger, cfg)
	if err != nil {
		return err
	}
	server := &http.Server{
		Addr:              cfg.HTTPAddress,
		Handler:           router,
		ReadTimeout:       readTimeout,
		ReadHeaderTimeout: readHeaderLimit,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
	}

	serverErrors := make(chan error, 1)
	go func() {
		logger.Info("API listening", "address", cfg.HTTPAddress)
		serverErrors <- server.ListenAndServe()
	}()

	select {
	case <-rootCtx.Done():
		logger.Info("shutdown signal received")
	case err := <-serverErrors:
		if !errors.Is(err, http.ErrServerClosed) {
			return err
		}
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	return server.Shutdown(shutdownCtx)
}

func routes(pool *pgxpool.Pool, logger *slog.Logger, cfg config.Config) (http.Handler, error) {
	catalogRepository := catalog.NewPostgresRepository(pool)
	catalogService := catalog.NewService(catalogRepository)
	catalogHandler := catalog.NewHandler(catalogService, logger)
	cartRepository := cart.NewPostgresRepository(pool)
	cartService := cart.NewService(cartRepository, catalogService)
	cartHandler := cart.NewHandler(cartService, logger)

	identityRepository := identity.NewPostgresRepository(pool)
	identityService, err := identity.NewService(
		identityRepository,
		identity.NewArgon2idHasher(identity.DefaultArgon2Parameters()),
		identity.NewCryptoSessionTokenGenerator(),
		cfg.SessionTTL,
	)
	if err != nil {
		return nil, err
	}
	identityHandler := identity.NewHandler(identityService, logger, identity.CookieConfig{
		Name: "aurora_session", Secure: cfg.CookieSecure,
	})

	router := chi.NewRouter()
	router.Use(httpx.RequestID)
	router.Use(httpx.AccessLog(logger))
	router.Use(httpx.Recoverer(logger))
	router.Get("/api/health/live", func(w http.ResponseWriter, _ *http.Request) {
		httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": "alive"})
	})
	router.Get("/api/health/ready", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		if err := pool.Ping(ctx); err != nil {
			httpx.WriteError(w, http.StatusServiceUnavailable, "not_ready", "Database is unavailable", nil)
			return
		}
		var schemaReady bool
		if err := pool.QueryRow(ctx, "SELECT to_regclass('public.products') IS NOT NULL AND to_regclass('public.skus') IS NOT NULL AND to_regclass('public.users') IS NOT NULL AND to_regclass('public.carts') IS NOT NULL AND to_regclass('public.cart_items') IS NOT NULL").Scan(&schemaReady); err != nil || !schemaReady {
			httpx.WriteError(w, http.StatusServiceUnavailable, "not_ready", "Database schema is not ready", nil)
			return
		}
		httpx.WriteJSON(w, http.StatusOK, map[string]string{"status": "ready"})
	})
	router.Get("/api/products", catalogHandler.ListProducts)
	router.Post("/api/products", catalogHandler.CreateProduct)
	router.Get("/api/products/{slug}", catalogHandler.GetProduct)
	router.With(httpx.RequireOrigin(cfg.PublicOrigin)).Post("/api/auth/register", identityHandler.Register)
	router.With(httpx.RequireOrigin(cfg.PublicOrigin)).Post("/api/auth/login", identityHandler.Login)
	router.With(httpx.RequireOrigin(cfg.PublicOrigin)).Post("/api/auth/logout", identityHandler.Logout)
	router.With(identityHandler.Authenticate).Get("/api/me", identityHandler.Me)
	router.With(identityHandler.Authenticate).Get("/api/auth/csrf", identityHandler.CSRFToken)
	router.With(identityHandler.Authenticate).Get("/api/cart", cartHandler.Get)
	router.With(identityHandler.Authenticate, httpx.RequireOrigin(cfg.PublicOrigin), identityHandler.RequireCSRF).Post("/api/cart/items", cartHandler.Add)
	router.With(identityHandler.Authenticate, httpx.RequireOrigin(cfg.PublicOrigin), identityHandler.RequireCSRF).Put("/api/cart/items/{skuID}", cartHandler.SetQuantity)
	router.With(identityHandler.Authenticate, httpx.RequireOrigin(cfg.PublicOrigin), identityHandler.RequireCSRF).Delete("/api/cart/items/{skuID}", cartHandler.Remove)
	return router, nil
}
