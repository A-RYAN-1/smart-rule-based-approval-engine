// Package config reads every setting from environment variables, in one
// place. The rest of the code uses the Config it returns. See .env.example
// for what each variable means.
package config

import (
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

// MinJWTSecretLength is the shortest JWT_SECRET accepted (32 bytes, the
// HS256 key size).
const MinJWTSecretLength = 32

type Config struct {
	// Port the HTTP server listens on (PORT, or APP_PORT, default 8080).
	Port string
	// DatabaseURL is a postgres:// connection string.
	DatabaseURL string
	// JWTSecret signs and verifies login tokens.
	JWTSecret string
	// CORSOrigins lists the browser origins allowed to call the API from
	// another domain. Empty means same-origin only.
	CORSOrigins []string
	// CookieSecure marks the login cookie HTTPS-only.
	CookieSecure bool
	// GinMode is "release" or "debug".
	GinMode string
	// CronSecret protects the HTTP trigger of the auto-reject job. Empty
	// disables the trigger.
	CronSecret string
	// RunCron starts the in-process nightly auto-reject job (long-running
	// server only; serverless hosts use the HTTP trigger instead).
	RunCron bool
	// Seed accounts, created on startup when the password is set and the
	// account does not exist yet.
	SeedAdminPassword   string
	SeedManagerPassword string
}

// Load reads .env (if present, for local runs) and then the environment.
// Variables already set in the environment win over .env.
func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		log.Printf("config: could not read .env: %v", err)
	}
	return FromEnv(os.Getenv)
}

// FromEnv builds and validates a Config from a lookup function, so tests
// can pass a map instead of the real environment.
func FromEnv(getenv func(string) string) (*Config, error) {
	get := func(key, def string) string {
		if v := strings.TrimSpace(getenv(key)); v != "" {
			return v
		}
		return def
	}

	cfg := &Config{
		Port:                get("PORT", get("APP_PORT", "8080")),
		JWTSecret:           getenv("JWT_SECRET"),
		CORSOrigins:         splitList(get("CORS_ORIGINS", "")),
		CookieSecure:        get("COOKIE_SECURE", "true") == "true",
		GinMode:             get("GIN_MODE", "release"),
		CronSecret:          getenv("CRON_SECRET"),
		RunCron:             get("RUN_CRON", "true") == "true",
		SeedAdminPassword:   getenv("SEED_ADMIN_PASSWORD"),
		SeedManagerPassword: getenv("SEED_MANAGER_PASSWORD"),
	}

	var problems []string

	cfg.DatabaseURL = get("DATABASE_URL", "")
	if cfg.DatabaseURL == "" {
		// Separate DB_* variables are still accepted for local setups.
		if host := get("DB_HOST", ""); host != "" {
			u := url.URL{
				Scheme:   "postgres",
				User:     url.UserPassword(get("DB_USER", "postgres"), getenv("DB_PASSWORD")),
				Host:     host + ":" + get("DB_PORT", "5432"),
				Path:     "/" + get("DB_NAME", "approval_engine"),
				RawQuery: "sslmode=" + get("DB_SSLMODE", "require"),
			}
			cfg.DatabaseURL = u.String()
		} else {
			problems = append(problems, "DATABASE_URL is not set (a postgres:// connection string)")
		}
	}

	if len(cfg.JWTSecret) < MinJWTSecretLength {
		problems = append(problems, fmt.Sprintf(
			"JWT_SECRET must be at least %d characters (generate one with: openssl rand -hex 32)",
			MinJWTSecretLength))
	}
	if cfg.GinMode != "release" && cfg.GinMode != "debug" && cfg.GinMode != "test" {
		problems = append(problems, "GIN_MODE must be release, debug or test")
	}
	for _, o := range cfg.CORSOrigins {
		if o == "*" {
			problems = append(problems, "CORS_ORIGINS cannot be * (the login cookie needs explicit origins)")
		}
	}
	for name, pw := range map[string]string{
		"SEED_ADMIN_PASSWORD":   cfg.SeedAdminPassword,
		"SEED_MANAGER_PASSWORD": cfg.SeedManagerPassword,
	} {
		if pw != "" && len(pw) < 8 {
			problems = append(problems, name+" must be at least 8 characters")
		}
	}

	if len(problems) > 0 {
		return nil, fmt.Errorf("invalid configuration:\n  - %s", strings.Join(problems, "\n  - "))
	}
	return cfg, nil
}

func splitList(s string) []string {
	var out []string
	for _, part := range strings.Split(s, ",") {
		if p := strings.TrimSpace(part); p != "" {
			out = append(out, strings.TrimSuffix(p, "/"))
		}
	}
	return out
}
