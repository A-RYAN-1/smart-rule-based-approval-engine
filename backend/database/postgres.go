package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/ankita-advitot/rule_based_approval_engine/config"
	"github.com/ankita-advitot/rule_based_approval_engine/interfaces"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var DB interfaces.DB

type pgxDB struct {
	pool *pgxpool.Pool
}

func (d *pgxDB) Begin(ctx context.Context) (interfaces.Tx, error) {
	tx, err := d.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	return &pgxTx{tx: tx}, nil
}

func (d *pgxDB) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	return d.pool.QueryRow(ctx, sql, args...)
}

func (d *pgxDB) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	return d.pool.Query(ctx, sql, args...)
}

func (d *pgxDB) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	return d.pool.Exec(ctx, sql, args...)
}

type pgxTx struct {
	tx pgx.Tx
}

func (t *pgxTx) Commit(ctx context.Context) error {
	return t.tx.Commit(ctx)
}

func (t *pgxTx) Rollback(ctx context.Context) error {
	return t.tx.Rollback(ctx)
}

func (t *pgxTx) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	return t.tx.QueryRow(ctx, sql, args...)
}

func (t *pgxTx) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	return t.tx.Query(ctx, sql, args...)
}

func (t *pgxTx) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	return t.tx.Exec(ctx, sql, args...)
}

// Connect opens the connection pool and checks it with a ping.
func Connect(ctx context.Context, cfg *config.Config) (*pgxpool.Pool, error) {
	poolCfg, err := pgxpool.ParseConfig(cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("DATABASE_URL is not a valid connection string: %w", err)
	}
	poolCfg.ConnConfig.ConnectTimeout = 10 * time.Second
	poolCfg.ConnConfig.RuntimeParams["application_name"] = "approval-engine"
	poolCfg.MaxConns = 5

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("unable to create database pool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("database ping failed: %w", err)
	}

	DB = &pgxDB{pool: pool}
	log.Println("PostgreSQL connected")
	return pool, nil
}
