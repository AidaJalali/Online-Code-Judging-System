package main

import (
	"flag"
	"fmt"
	"log"
	"path/filepath"

	"online-judge/internal/config"
	"online-judge/internal/database"
)

func main() {
	// Parse command line flags
	configPath := flag.String("config", "config.yaml", "path to config file")
	flag.Parse()

	// Load configuration
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	// Initialize database connection
	db, err := database.NewDB(cfg.Database)
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}
	defer db.Close()

	// Get the absolute path to the migrations directory
	migrationsDir, err := filepath.Abs("migrations")
	if err != nil {
		log.Fatalf("Error getting migrations directory path: %v", err)
	}

	// Run migrations
	if err := database.RunMigrations(db, migrationsDir); err != nil {
		log.Fatalf("Error running migrations: %v", err)
	}

	// TODO: Initialize and start the HTTP server
	fmt.Println("Database migrations completed successfully")
}
