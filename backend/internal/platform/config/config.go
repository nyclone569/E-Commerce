package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Environment     string
	Version         string
	HTTPAddress     string
	DatabaseURL     string
	DatabaseMaxOpen int32
	DatabaseMinIdle int32
	DatabaseTimeout time.Duration
	PublicOrigin    string
	SessionTTL      time.Duration
	CookieSecure    bool
}

func Load() (Config, error) {
	cfg := Config{
		Environment:     envOrDefault("APP_ENV", "development"),
		Version:         envOrDefault("APP_VERSION", "dev"),
		HTTPAddress:     envOrDefault("HTTP_ADDRESS", ":8080"),
		DatabaseURL:     os.Getenv("DATABASE_URL"),
		DatabaseMaxOpen: 10,
		DatabaseMinIdle: 2,
		DatabaseTimeout: 5 * time.Second,
		PublicOrigin:    envOrDefault("PUBLIC_ORIGIN", "http://localhost:3000"),
		SessionTTL:      24 * time.Hour,
	}
	cfg.CookieSecure = cfg.Environment != "development" && cfg.Environment != "test"
	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}

	var err error
	if cfg.DatabaseMaxOpen, err = int32Env("DATABASE_MAX_OPEN_CONNS", cfg.DatabaseMaxOpen); err != nil {
		return Config{}, err
	}
	if cfg.DatabaseMinIdle, err = int32Env("DATABASE_MIN_IDLE_CONNS", cfg.DatabaseMinIdle); err != nil {
		return Config{}, err
	}
	if cfg.DatabaseMinIdle > cfg.DatabaseMaxOpen {
		return Config{}, fmt.Errorf("DATABASE_MIN_IDLE_CONNS must not exceed DATABASE_MAX_OPEN_CONNS")
	}
	if cfg.DatabaseMaxOpen == 0 {
		return Config{}, fmt.Errorf("DATABASE_MAX_OPEN_CONNS must be greater than zero")
	}
	if raw := os.Getenv("DATABASE_CONNECT_TIMEOUT"); raw != "" {
		cfg.DatabaseTimeout, err = time.ParseDuration(raw)
		if err != nil {
			return Config{}, fmt.Errorf("parse DATABASE_CONNECT_TIMEOUT: %w", err)
		}
	}
	if raw := os.Getenv("SESSION_TTL"); raw != "" {
		cfg.SessionTTL, err = time.ParseDuration(raw)
		if err != nil || cfg.SessionTTL <= 0 || cfg.SessionTTL > 30*24*time.Hour {
			return Config{}, fmt.Errorf("SESSION_TTL must be a positive duration no greater than 720h")
		}
	}
	if cfg.CookieSecure, err = boolEnv("SESSION_COOKIE_SECURE", cfg.CookieSecure); err != nil {
		return Config{}, err
	}
	if cfg.Environment == "production" && !cfg.CookieSecure {
		return Config{}, fmt.Errorf("SESSION_COOKIE_SECURE must be true in production")
	}
	if err := validateOrigin(cfg.PublicOrigin); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func int32Env(key string, fallback int32) (int32, error) {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback, nil
	}
	value, err := strconv.ParseInt(raw, 10, 32)
	if err != nil || value < 0 {
		return 0, fmt.Errorf("%s must be a non-negative integer", key)
	}
	return int32(value), nil
}

func boolEnv(key string, fallback bool) (bool, error) {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback, nil
	}
	value, err := strconv.ParseBool(raw)
	if err != nil {
		return false, fmt.Errorf("%s must be a boolean", key)
	}
	return value, nil
}

func validateOrigin(rawOrigin string) error {
	origin, err := url.Parse(rawOrigin)
	if err != nil || (origin.Scheme != "http" && origin.Scheme != "https") || origin.Host == "" || origin.User != nil || origin.Path != "" || origin.RawQuery != "" || origin.Fragment != "" {
		return fmt.Errorf("PUBLIC_ORIGIN must contain only an http or https scheme and host")
	}
	return nil
}
