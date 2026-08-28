package config

import (
	"testing"
	"time"
)

func TestLoadRequiresDatabaseURL(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	if _, err := Load(); err == nil {
		t.Fatal("Load() error = nil, want required DATABASE_URL error")
	}
}

func TestLoadRejectsInvalidPoolBudget(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://example")
	t.Setenv("DATABASE_MAX_OPEN_CONNS", "2")
	t.Setenv("DATABASE_MIN_IDLE_CONNS", "3")
	if _, err := Load(); err == nil {
		t.Fatal("Load() error = nil, want pool budget error")
	}
}

func TestLoadUsesExplicitConfiguration(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://example")
	t.Setenv("DATABASE_MAX_OPEN_CONNS", "12")
	t.Setenv("DATABASE_MIN_IDLE_CONNS", "4")
	t.Setenv("DATABASE_CONNECT_TIMEOUT", "3s")
	t.Setenv("PUBLIC_ORIGIN", "https://shop.example.com")
	t.Setenv("SESSION_TTL", "12h")
	t.Setenv("SESSION_COOKIE_SECURE", "true")
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.DatabaseMaxOpen != 12 || cfg.DatabaseMinIdle != 4 || cfg.DatabaseTimeout.String() != "3s" || cfg.SessionTTL != 12*time.Hour || !cfg.CookieSecure {
		t.Fatalf("unexpected config: %#v", cfg)
	}
}

func TestLoadRejectsInvalidIdentityConfiguration(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://example")
	t.Setenv("PUBLIC_ORIGIN", "https://shop.example.com/path")
	if _, err := Load(); err == nil {
		t.Fatal("Load() error = nil, want invalid origin error")
	}

	t.Setenv("PUBLIC_ORIGIN", "https://shop.example.com")
	t.Setenv("SESSION_TTL", "0s")
	if _, err := Load(); err == nil {
		t.Fatal("Load() error = nil, want invalid session TTL error")
	}
}

func TestLoadRejectsInsecureProductionCookie(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://example")
	t.Setenv("APP_ENV", "production")
	t.Setenv("SESSION_COOKIE_SECURE", "false")
	if _, err := Load(); err == nil {
		t.Fatal("Load() error = nil, want insecure production cookie error")
	}
}
