package database

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// Migration represents a database migration
type Migration struct {
	Version int
	Up      string
	Down    string
}

// RunMigrations executes all pending migrations
func RunMigrations(db *sqlx.DB, migrationsDir string) error {
	// Create migrations table if it doesn't exist
	if err := createMigrationsTable(db); err != nil {
		return fmt.Errorf("error creating migrations table: %w", err)
	}

	// Get all migration files
	files, err := ioutil.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("error reading migrations directory: %w", err)
	}

	// Parse migration files
	migrations := make([]Migration, 0)
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".up.sql") {
			version := extractVersion(file.Name())
			upPath := filepath.Join(migrationsDir, file.Name())
			downPath := filepath.Join(migrationsDir, fmt.Sprintf("%d_init_schema.down.sql", version))

			upSQL, err := ioutil.ReadFile(upPath)
			if err != nil {
				return fmt.Errorf("error reading up migration %s: %w", upPath, err)
			}

			downSQL, err := ioutil.ReadFile(downPath)
			if err != nil {
				return fmt.Errorf("error reading down migration %s: %w", downPath, err)
			}

			migrations = append(migrations, Migration{
				Version: version,
				Up:      string(upSQL),
				Down:    string(downSQL),
			})
		}
	}

	// Sort migrations by version
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	// Get current version
	currentVersion, err := getCurrentVersion(db)
	if err != nil {
		return fmt.Errorf("error getting current version: %w", err)
	}

	// Run pending migrations
	for _, migration := range migrations {
		if migration.Version > currentVersion {
			if err := runMigration(db, migration); err != nil {
				return fmt.Errorf("error running migration %d: %w", migration.Version, err)
			}
		}
	}

	return nil
}

func createMigrationsTable(db *sqlx.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS migrations (
			version INTEGER PRIMARY KEY,
			applied_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		)
	`)
	return err
}

func getCurrentVersion(db *sqlx.DB) (int, error) {
	var version sql.NullInt64
	err := db.Get(&version, "SELECT MAX(version) FROM migrations")
	if err != nil {
		return 0, err
	}
	if !version.Valid {
		return 0, nil
	}
	return int(version.Int64), nil
}

func runMigration(db *sqlx.DB, migration Migration) error {
	tx, err := db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Run the migration
	if _, err := tx.Exec(migration.Up); err != nil {
		return fmt.Errorf("error executing up migration: %w", err)
	}

	// Record the migration
	if _, err := tx.Exec("INSERT INTO migrations (version) VALUES ($1)", migration.Version); err != nil {
		return fmt.Errorf("error recording migration: %w", err)
	}

	return tx.Commit()
}

func extractVersion(filename string) int {
	var version int
	fmt.Sscanf(filename, "%d_", &version)
	return version
}
