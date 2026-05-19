package database

import (
	"fmt"
	"os"

	"github.com/floci-next-go/back/internal/infra/dto"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// OpenFromEnv connects with DATABASE_URL, or with PGHOST/PGPORT/PGUSER/PGPASSWORD/PGDATABASE
// (defaults align with the docker-compose `database` service).
func OpenFromEnv() (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(PostgresDSNFromEnv()), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	if err := db.AutoMigrate(
		&dto.Todo{},
		&dto.File{},
		&dto.TodosFiles{},
	); err != nil {
		return nil, err
	}
	return db, nil
}

// PostgresDSNFromEnv returns DATABASE_URL when set; otherwise a libpq-style DSN from PG* variables.
func PostgresDSNFromEnv() string {
	if u := os.Getenv("DATABASE_URL"); u != "" {
		return u
	}
	host := getenvDefault("PGHOST", "localhost")
	port := getenvDefault("PGPORT", "5432")
	user := getenvDefault("PGUSER", "root")
	pass := getenvDefault("PGPASSWORD", "root")
	dbname := getenvDefault("PGDATABASE", "db")
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable TimeZone=UTC",
		host, port, user, pass, dbname,
	)
}

// PostgresTargetSummary is for logs only (no password).
func PostgresTargetSummary() string {
	if os.Getenv("DATABASE_URL") != "" {
		return "DATABASE_URL"
	}
	return fmt.Sprintf(
		"host=%s port=%s db=%s user=%s",
		getenvDefault("PGHOST", "localhost"),
		getenvDefault("PGPORT", "5432"),
		getenvDefault("PGDATABASE", "db"),
		getenvDefault("PGUSER", "root"),
	)
}

func getenvDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
