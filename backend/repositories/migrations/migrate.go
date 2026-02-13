package migrations

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/ankita-advitot/rule_based_approval_engine/database"
	"github.com/ankita-advitot/rule_based_approval_engine/pkg/utils"
)

const migrationsDir = "migrations"

type Migration struct {
	directoryName string
}

func createMigration() {
	if len(os.Args) < 4 {
		log.Println("migration name required")
		return
	}

	m := Migration{directoryName: migrationsDir}
	err := m.CreateMigrationFile(os.Args[3])
	if err != nil {
		log.Println(err)
	}
}

func (m Migration) CreateMigrationFile(name string) error {
	if name == "" {
		return errors.New("filename is not provided")
	}

	ts := time.Now().Unix()
	up := fmt.Sprintf("%s/%d_%s.up.sql", m.directoryName, ts, name)
	down := fmt.Sprintf("%s/%d_%s.down.sql", m.directoryName, ts, name)

	os.MkdirAll(m.directoryName, 0755)
	os.Create(up)
	os.Create(down)

	log.Println("created", up)
	log.Println("created", down)
	return nil
}

func HandleMigrateCommand() {
	if len(os.Args) < 3 {
		log.Println("migrate command required (create | up)")
		return
	}
	log.Println("in handlemigrate", os.Args[2])
	switch os.Args[2] {
	case "create":
		createMigration()
	case "up":
		log.Println("calling to run")
		RunMigrationsUp()
	default:
		log.Println("unknown migrate command")
	}
}
func RunMigrationsUp() {
	if database.DB == nil {
		log.Fatal("Database connection not initialized")
	}
	log.Println("in run")

	ensureSchemaTable()
	log.Println("calling to read")
	files := readUpSQLFiles()

	for _, file := range files {
		version := extractVersion(file)

		if isApplied(version) {
			continue
		}

		log.Println("applying", file)
		err := executeSQLFile(file)
		if err != nil {
			log.Fatal(err)
		}

		markAsApplied(version)
	}
}

func ensureSchemaTable() {
	log.Println("in ensure")

	_, err := database.DB.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version BIGINT PRIMARY KEY,
			applied_at TIMESTAMP NOT NULL
		)
	`)
	log.Println("after query")

	if err != nil {
		log.Fatalf("Error ensuring schema table: %v", err)
	}
}

func readUpSQLFiles() []string {
	log.Println("in read")
	entries, _ := os.ReadDir(migrationsDir)
	var files []string
	log.Println("after read")

	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".up.sql") {
			files = append(files, migrationsDir+"/"+e.Name())
		}
	}
	log.Println("files", files)
	sort.Strings(files)
	return files
}

func extractVersion(path string) int64 {
	base := filepath.Base(path)
	parts := strings.Split(base, "_")
	v, _ := strconv.ParseInt(parts[0], 10, 64)
	return v
}

func executeSQLFile(path string) error {
	log.Println("in execute")
	sqlBytes, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	log.Println("after read")

	_, err = database.DB.Exec(context.Background(), string(sqlBytes))
	log.Println("after exec", err)
	return utils.MapPgError(err)
}

func isApplied(version int64) bool {
	var v int64
	err := database.DB.QueryRow(
		context.Background(),
		"SELECT version FROM schema_migrations WHERE version=$1",
		version,
	).Scan(&v)
	return err == nil
}

func markAsApplied(version int64) {
	log.Println("in mark")
	_, err := database.DB.Exec(
		context.Background(),
		"INSERT INTO schema_migrations (version, applied_at) VALUES ($1, NOW())",
		version,
	)
	if err != nil {
		log.Printf("Warning: Failed to mark migration %d as applied: %v", version, err)
	}
	log.Println("after mark")
}
