// Package migrations applies the SQL files embedded from backend/migrations,
// each in its own transaction, and records them in schema_migrations.
package migrations

import (
	"context"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/ankita-advitot/rule_based_approval_engine/interfaces"
	migrationfiles "github.com/ankita-advitot/rule_based_approval_engine/migrations"
)

// HandleMigrateCommand runs `server migrate up` or `server migrate create <name>`.
func HandleMigrateCommand(ctx context.Context, db interfaces.DB, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: server migrate up | server migrate create <name>")
	}
	switch args[0] {
	case "up":
		return Up(ctx, db)
	case "create":
		if len(args) < 2 || strings.TrimSpace(args[1]) == "" {
			return fmt.Errorf("usage: server migrate create <name>")
		}
		return createFiles("migrations", args[1])
	default:
		return fmt.Errorf("unknown migrate command %q (use up or create)", args[0])
	}
}

// createFiles writes an empty up/down pair into dir (run from backend/).
func createFiles(dir, name string) error {
	ts := time.Now().Unix()
	for _, suffix := range []string{"up", "down"} {
		path := filepath.Join(dir, fmt.Sprintf("%d_%s.%s.sql", ts, name, suffix))
		if err := os.WriteFile(path, nil, 0o644); err != nil {
			return err
		}
		log.Println("created", path)
	}
	return nil
}

// Up applies every embedded *.up.sql file that is not applied yet.
func Up(ctx context.Context, db interfaces.DB) error {
	if _, err := db.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version BIGINT PRIMARY KEY,
			applied_at TIMESTAMP NOT NULL
		)`); err != nil {
		return fmt.Errorf("creating schema_migrations: %w", err)
	}

	files, err := upFiles()
	if err != nil {
		return err
	}

	for _, file := range files {
		version, err := versionOf(file)
		if err != nil {
			return err
		}
		var applied bool
		if err := db.QueryRow(ctx,
			"SELECT EXISTS (SELECT 1 FROM schema_migrations WHERE version = $1)", version,
		).Scan(&applied); err != nil {
			return fmt.Errorf("checking migration %s: %w", file, err)
		}
		if applied {
			continue
		}

		sql, err := fs.ReadFile(migrationfiles.FS, file)
		if err != nil {
			return err
		}
		if err := apply(ctx, db, version, string(sql)); err != nil {
			return fmt.Errorf("applying migration %s: %w", file, err)
		}
		log.Println("applied migration", file)
	}
	return nil
}

func apply(ctx context.Context, db interfaces.DB, version int64, sql string) error {
	tx, err := db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, sql); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx,
		"INSERT INTO schema_migrations (version, applied_at) VALUES ($1, NOW())", version,
	); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func upFiles() ([]string, error) {
	entries, err := fs.ReadDir(migrationfiles.FS, ".")
	if err != nil {
		return nil, err
	}
	var files []string
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".up.sql") {
			files = append(files, e.Name())
		}
	}
	sort.Strings(files)
	return files, nil
}

func versionOf(file string) (int64, error) {
	prefix, _, _ := strings.Cut(file, "_")
	v, err := strconv.ParseInt(prefix, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("migration %s: name must start with a unix timestamp", file)
	}
	return v, nil
}
