package config

import (
	"mamabloemetjes_server/structs"
	"testing"
	"time"
)

func validTestConfig() *structs.Config {
	return &structs.Config{
		Server: &structs.ServerConfig{
			Environment: "production", LogLevel: "info", MonitoringToken: "12345678901234567890123456789012",
		},
		Database: &structs.DatabaseConfig{MinConns: 1, MaxConns: 2},
		Auth: &structs.AuthConfig{
			AccessTokenExpiry: 15 * time.Minute, RefreshTokenExpiry: 24 * time.Hour,
		},
		Cache: &structs.CacheConfig{
			MinIdleConns: 1, MaxIdleConns: 2, PoolSize: 3,
			MinRetryBackoff: time.Millisecond, MaxRetryBackoff: time.Second,
		},
		RateLimit: &structs.RateLimitConfig{Enabled: true},
	}
}

func TestValidateConfigAcceptsSecureProductionConfig(t *testing.T) {
	if err := validateConfig(validTestConfig()); err != nil {
		t.Fatalf("validateConfig() error = %v", err)
	}
}

func TestValidateConfigRejectsInsecureProductionConfig(t *testing.T) {
	cfg := validTestConfig()
	cfg.Server.LogLevel = "debug"
	if err := validateConfig(cfg); err == nil {
		t.Fatal("expected debug logging to be rejected in production")
	}

	cfg = validTestConfig()
	cfg.RateLimit.Enabled = false
	if err := validateConfig(cfg); err == nil {
		t.Fatal("expected disabled rate limiting to be rejected in production")
	}

	cfg = validTestConfig()
	cfg.Server.MonitoringToken = ""
	if err := validateConfig(cfg); err == nil {
		t.Fatal("expected missing monitoring token to be rejected in production")
	}
}
