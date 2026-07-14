package config

import (
	"testing"
)

func TestLoadRequiresDatabaseURL(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	t.Setenv("AUTH_JWT_SECRET", "secret")

	_, err := Load()
	if err == nil || err.Error() != "DATABASE_URL is required" {
		t.Fatalf("expected DATABASE_URL validation error, got %v", err)
	}
}

func TestLoadRequiresAuthJWTSecret(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://example")
	t.Setenv("AUTH_JWT_SECRET", "")

	_, err := Load()
	if err == nil || err.Error() != "AUTH_JWT_SECRET is required" {
		t.Fatalf("expected AUTH_JWT_SECRET validation error, got %v", err)
	}
}

func TestLoadReturnsConfigWhenRequiredEnvVarsExist(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://example")
	t.Setenv("AUTH_JWT_SECRET", "secret")
	t.Setenv("PORT", "9090")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected config to load, got error %v", err)
	}

	if cfg.DatabaseURL != "postgres://example" {
		t.Fatalf("expected database url to be loaded, got %q", cfg.DatabaseURL)
	}

	if cfg.AuthJWTSecret != "secret" {
		t.Fatalf("expected jwt secret to be loaded, got %q", cfg.AuthJWTSecret)
	}

	if cfg.Port != "9090" {
		t.Fatalf("expected port to be loaded, got %q", cfg.Port)
	}
}