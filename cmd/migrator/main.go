package main

import (
	"auth-api/internal/config"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	cfg := config.MustLoad()

	db := cfg.Database

	migrationsPath := "file://migrations"
	migrationsTable := "schema_migrations"

	dbURL := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=disable&x-migrations-table=%s",
		db.User, db.Password, db.Host, db.Port, db.Name, migrationsTable,
	)

	m, err := migrate.New(migrationsPath, dbURL)
	if err != nil {
		log.Fatalf("failed to create migrator: %v", err)
	}

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			fmt.Println("✅ No new migrations.")
			os.Exit(0)
		}
		log.Fatalf("❌ Migration failed: %v", err)
	}

	fmt.Println("✅ Migrations applied successfully.")
}
