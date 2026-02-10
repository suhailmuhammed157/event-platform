package storage

import (
	"errors"
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	// Import the PostgreSQL driver for database connections
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	// Import the file source driver to read migrations from filesystem
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// RunMigrations applies all pending database migrations
// It returns an error if migration fails (except for "no change" which is not an error)
func RunMigrations(databaseURL, migrationsPath string) error {
	// Create a new migrate instance with the migrations path and database URL
	m, err := migrate.New(
		fmt.Sprintf("file://%s", migrationsPath),
		databaseURL,
	)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}
	defer m.Close()

	// Apply all pending migrations
	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	// Log the current version after migration
	version, dirty, err := m.Version()
	if err != nil && !errors.Is(err, migrate.ErrNilVersion) {
		return fmt.Errorf("failed to get migration version: %w", err)
	}

	if dirty {
		log.Printf("Warning: Database is in dirty state at version %d", version)
	} else {
		log.Printf("Database migrated to version %d", version)
	}

	return nil
}
