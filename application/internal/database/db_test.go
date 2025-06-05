package database

import (
	"testing"
)

func TestDatabaseConnection(t *testing.T) {
	// Use hardcoded config for local Postgres
	cfg := Config{
		Host:            "localhost",
		Port:            5432,
		User:            "mahdi",
		Password:        "secret123",
		DBName:          "online-judge",
		SSLMode:         "disable",
		MaxOpenConns:    25,
		MaxIdleConns:    5,
		ConnMaxLifetime: 0, // or 5 * time.Minute if needed
	}

	// Initialize database connection
	db, err := NewDB(cfg)
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Test a simple query
	var result int
	err = db.Get(&result, "SELECT 1")
	if err != nil {
		t.Fatalf("Failed to execute test query: %v", err)
	}

	if result != 1 {
		t.Errorf("Expected 1, got %d", result)
	}

	t.Log("Database connection test passed successfully")
}
