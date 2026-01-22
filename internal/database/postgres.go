package database

import (
	"context"
	"fmt"
	"log"

	"rule-based-approval-engine/internal/config"

	"github.com/jackc/pgx/v5/pgxpool"
)

var DB *pgxpool.Pool

func Connect(cfg *config.Config) {
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.DB.User,
		cfg.DB.Password,
		cfg.DB.Host,
		cfg.DB.Port,
		cfg.DB.Name,
		cfg.DB.SSLMode,
	)

	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		log.Fatal("Unable to connect to database:", err)
	}

	err = pool.Ping(context.Background())
	if err != nil {
		log.Fatal("Database ping failed:", err)
	}

	DB = pool
	log.Println("✅ PostgreSQL connected")
}
