package database

import (
	"path/filepath"
	"runtime"
	"testing"

	"online-judge/internal/config"
)

func TestDatabaseConnection(t *testing.T) {
	// Get the absolute path to the config file
	_, currentFile, _, _ := runtime.Caller(0)
	configPath := filepath.Join(filepath.Dir(filepath.Dir(filepath.Dir(currentFile))), "configs", "config.yaml")

	// Load configuration
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Initialize database connection
	db, err := NewDB(Config{
		Host:            cfg.Database.Host,
		Port:            cfg.Database.Port,
		User:            cfg.Database.User,
		Password:        cfg.Database.Password,
		DBName:          cfg.Database.DBName,
		SSLMode:         cfg.Database.SSLMode,
		MaxOpenConns:    cfg.Database.MaxOpenConns,
		MaxIdleConns:    cfg.Database.MaxIdleConns,
		ConnMaxLifetime: cfg.Database.ConnMaxLifetime,
	})
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
