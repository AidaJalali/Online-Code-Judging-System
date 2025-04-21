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

	log.Println("Successfully connected to the database!")

	// Test query
	rows, err := db.Query("SELECT 1")
	if err != nil {
		log.Fatalf("Failed to execute test query: %v", err)
	}
	defer rows.Close()

	log.Println("Successfully executed test query!")
}
