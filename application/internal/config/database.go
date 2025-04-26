package config

import (
	"database/sql"
	"fmt"
	"sync"
	"time"

	_ "github.com/lib/pq"
)

var (
	db     *sql.DB
	dbOnce sync.Once
)

// InitDB initializes the database connection pool
func InitDB() (*sql.DB, error) {
	var err error
	dbOnce.Do(func() {
		// Replace these with your actual database configuration
		connStr := "user=postgres dbname=online_judge sslmode=disable password=postgres host=localhost port=5432"

		db, err = sql.Open("postgres", connStr)
		if err != nil {
			return
		}

		// Set connection pool settings
		db.SetMaxOpenConns(25)
		db.SetMaxIdleConns(25)
		db.SetConnMaxLifetime(5 * time.Minute)

		// Test the connection
		err = db.Ping()
	})

	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	return db, nil
}

// GetDB returns the database connection pool
func GetDB() *sql.DB {
	return db
}

// CloseDB closes the database connection pool
func CloseDB() error {
	if db != nil {
		return db.Close()
	}
	return nil
}
