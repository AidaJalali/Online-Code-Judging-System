package main

import (
	"fmt"
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

	// Print connection details (without password)
	fmt.Println("Attempting to connect with these settings:")
	fmt.Printf("Host: %s\n", cfg.Database.Host)
	fmt.Printf("Port: %d\n", cfg.Database.Port)
	fmt.Printf("User: %s\n", cfg.Database.User)
	fmt.Printf("Database: %s\n", cfg.Database.DBName)
	fmt.Printf("SSL Mode: %s\n", cfg.Database.SSLMode)

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
		ConnectTimeout:  5,
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Test connection with a simple query
	var version string
	err = db.Get(&version, "SELECT version()")
	if err != nil {
		log.Fatalf("Failed to get PostgreSQL version: %v", err)
	}
	fmt.Printf("\nSuccessfully connected to PostgreSQL!\nVersion: %s\n", version)

	// Check current user and database
	var currentUser, currentDB string
	err = db.Get(&currentUser, "SELECT current_user")
	if err != nil {
		log.Printf("Warning: Could not get current user: %v", err)
	} else {
		fmt.Printf("Current user: %s\n", currentUser)
	}

	err = db.Get(&currentDB, "SELECT current_database()")
	if err != nil {
		log.Printf("Warning: Could not get current database: %v", err)
	} else {
		fmt.Printf("Current database: %s\n", currentDB)
	}

	// List available databases
	var databases []string
	err = db.Select(&databases, "SELECT datname FROM pg_database WHERE datistemplate = false")
	if err != nil {
		log.Printf("Warning: Could not list databases: %v", err)
	} else {
		fmt.Println("\nAvailable databases:")
		for _, dbname := range databases {
			fmt.Printf("- %s\n", dbname)
		}
	}
}
