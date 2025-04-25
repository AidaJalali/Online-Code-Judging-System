package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

const (
	host     = "b7cc3bd2-3f82-461a-b79f-21b3dd0b7461.hadb.ir"
	port     = 30226
	user     = "postgres"
	password = "feJ9hH6xiSBkT27G4PW5"
	dbname   = "postgres"
)

func main() {
	// Print connection details (without password)
	fmt.Println("Attempting to connect with these settings:")
	fmt.Printf("Host: %s\n", host)
	fmt.Printf("Port: %d\n", port)
	fmt.Printf("User: %s\n", user)
	fmt.Printf("Database: %s\n", dbname)

	// Initialize database connection
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatalf("Failed to open database connection: %v", err)
	}
	defer db.Close()

	// Test connection
	err = db.Ping()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Test connection with a simple query
	var version string
	err = db.QueryRow("SELECT version()").Scan(&version)
	if err != nil {
		log.Fatalf("Failed to get PostgreSQL version: %v", err)
	}
	fmt.Printf("\nSuccessfully connected to PostgreSQL!\nVersion: %s\n", version)

	// Check current user and database
	var currentUser, currentDB string
	err = db.QueryRow("SELECT current_user").Scan(&currentUser)
	if err != nil {
		log.Printf("Warning: Could not get current user: %v", err)
	} else {
		fmt.Printf("Current user: %s\n", currentUser)
	}

	err = db.QueryRow("SELECT current_database()").Scan(&currentDB)
	if err != nil {
		log.Printf("Warning: Could not get current database: %v", err)
	} else {
		fmt.Printf("Current database: %s\n", currentDB)
	}

	// List available tables
	rows, err := db.Query(`
		SELECT table_name 
		FROM information_schema.tables 
		WHERE table_schema = 'public'
	`)
	if err != nil {
		log.Printf("Warning: Could not list tables: %v", err)
	} else {
		fmt.Println("\nAvailable tables:")
		for rows.Next() {
			var tableName string
			if err := rows.Scan(&tableName); err != nil {
				log.Printf("Error scanning table name: %v", err)
				continue
			}
			fmt.Printf("- %s\n", tableName)
		}
		rows.Close()
	}

	// Check if users table exists and has data
	var userCount int
	err = db.QueryRow("SELECT COUNT(*) FROM users").Scan(&userCount)
	if err != nil {
		log.Printf("Warning: Could not count users: %v", err)
	} else {
		fmt.Printf("\nNumber of users in database: %d\n", userCount)
	}
}
