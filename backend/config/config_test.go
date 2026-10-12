package config

import (
	"strings"
	"testing"
)

const secret = "0123456789abcdef0123456789abcdef"

func env(m map[string]string) func(string) string {
	return func(k string) string { return m[k] }
}

func TestFromEnvDefaults(t *testing.T) {
	cfg, err := FromEnv(env(map[string]string{
		"DATABASE_URL": "postgres://u:p@localhost:5432/db",
		"JWT_SECRET":   secret,
	}))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Port != "8080" || cfg.GinMode != "release" || !cfg.CookieSecure || !cfg.RunCron {
		t.Errorf("unexpected defaults: %+v", cfg)
	}
	if len(cfg.CORSOrigins) != 0 {
		t.Errorf("CORS should default to same-origin only, got %v", cfg.CORSOrigins)
	}
}

func TestFromEnvPortPrefersPORT(t *testing.T) {
	cfg, err := FromEnv(env(map[string]string{
		"DATABASE_URL": "postgres://x", "JWT_SECRET": secret, "PORT": "9000", "APP_PORT": "7000",
	}))
	if err != nil || cfg.Port != "9000" {
		t.Fatalf("got %v, %v", cfg, err)
	}
}

func TestFromEnvBuildsURLFromDBVars(t *testing.T) {
	cfg, err := FromEnv(env(map[string]string{
		"JWT_SECRET": secret, "DB_HOST": "localhost", "DB_PORT": "5433",
		"DB_USER": "approval", "DB_PASSWORD": "p@ss word", "DB_NAME": "approval_engine", "DB_SSLMODE": "disable",
	}))
	if err != nil {
		t.Fatal(err)
	}
	want := "postgres://approval:p%40ss%20word@localhost:5433/approval_engine?sslmode=disable"
	if cfg.DatabaseURL != want {
		t.Errorf("DatabaseURL = %q, want %q", cfg.DatabaseURL, want)
	}
}

func TestFromEnvRejectsBadConfig(t *testing.T) {
	_, err := FromEnv(env(map[string]string{
		"JWT_SECRET":          "short",
		"CORS_ORIGINS":        "*",
		"SEED_ADMIN_PASSWORD": "123",
	}))
	if err == nil {
		t.Fatal("expected an error")
	}
	for _, want := range []string{"DATABASE_URL", "JWT_SECRET", "CORS_ORIGINS", "SEED_ADMIN_PASSWORD"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error should mention %s: %v", want, err)
		}
	}
}

func TestFromEnvSplitsCORSOrigins(t *testing.T) {
	cfg, err := FromEnv(env(map[string]string{
		"DATABASE_URL": "postgres://x", "JWT_SECRET": secret,
		"CORS_ORIGINS": "http://localhost:5173/, https://example.com ,",
	}))
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.CORSOrigins) != 2 || cfg.CORSOrigins[0] != "http://localhost:5173" || cfg.CORSOrigins[1] != "https://example.com" {
		t.Errorf("CORSOrigins = %v", cfg.CORSOrigins)
	}
}
