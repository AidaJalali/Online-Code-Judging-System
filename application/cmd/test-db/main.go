package main

import (
	"log"
	"path/filepath"
	"runtime"

	"online-judge/internal/config"
	"online-judge/internal/database"
)

func main() {
	// Get the absolute path to the config file
	_, currentFile, _, _ := runtime.Caller(0)
	configPath := filepath.Join(filepath.Dir(filepath.Dir(filepath.Dir(currentFile))), "configs", "config.yaml")

	// Load configuration
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize database connection
	db, err := database.NewDB(database.Config{
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
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Test a simple query
	var result int
	err = db.Get(&result, "SELECT 1")
	if err != nil {
		log.Fatalf("Failed to execute test query: %v", err)
	}

	log.Printf("Database connection test successful! Result: %d", result)

	// Test listing databases (requires superuser privileges)
	var databases []string
	err = db.Select(&databases, "SELECT datname FROM pg_database")
	if err != nil {
		log.Printf("Warning: Could not list databases (might need superuser privileges): %v", err)
	} else {
		log.Println("Available databases:")
		for _, dbname := range databases {
			log.Printf("- %s", dbname)
		}
	}
}
