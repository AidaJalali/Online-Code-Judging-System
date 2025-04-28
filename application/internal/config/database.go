package config

import (
	"database/sql"
	"fmt"
	"time"

	"online-judge/internal/logger"

	_ "github.com/lib/pq"
)

const (
	host     = "82eab2ea-bbbc-4025-9858-5fa2738ec7e4.hsvc.ir"
	port     = 32215
	user     = "postgres"
	password = "KwuujJsBtnS4cTbowdpZYjdlKy8Vk0Du"
	dbname   = "postgres"
)

func InitDB() (*sql.DB, error) {
	logger.Info("Initializing database connection...")
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	logger.Info("Opening database connection...")
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		logger.Error("Failed to open database: %v", err)
		return nil, fmt.Errorf("error opening database: %v", err)
	}

	logger.Info("Pinging database...")
	err = db.Ping()
	if err != nil {
		logger.Error("Failed to ping database: %v", err)
		return nil, fmt.Errorf("error connecting to database: %v", err)
	}
	logger.Info("Database connection successful")

	// Test query to verify access
	logger.Info("Testing database access...")
	var testResult int
	err = db.QueryRow("SELECT 1").Scan(&testResult)
	if err != nil {
		logger.Error("Failed to execute test query: %v", err)
		return nil, fmt.Errorf("error testing database access: %v", err)
	}
	logger.Info("Database access test successful")

	// Create tables if they don't exist
	logger.Info("Creating tables if they don't exist...")
	err = createTables(db)
	if err != nil {
		logger.Error("Failed to create tables: %v", err)
		return nil, fmt.Errorf("error creating tables: %v", err)
	}
	logger.Info("Tables created successfully")

	// Insert test data
	logger.Info("Inserting test data...")
	err = insertTestData(db)
	if err != nil {
		logger.Error("Error inserting test data: %v", err)
	} else {
		logger.Info("Test data inserted successfully")
	}

	// Verify test data
	logger.Info("Verifying test data...")
	var questionCount int
	err = db.QueryRow("SELECT COUNT(*) FROM questions").Scan(&questionCount)
	if err != nil {
		logger.Error("Failed to count questions: %v", err)
	} else {
		logger.Info("Found %d questions in database", questionCount)
	}

	var userCount int
	err = db.QueryRow("SELECT COUNT(*) FROM users").Scan(&userCount)
	if err != nil {
		logger.Error("Failed to count users: %v", err)
	} else {
		logger.Info("Found %d users in database", userCount)
	}

	return db, nil
}

func createTables(db *sql.DB) error {
	// Create users table
	logger.Info("Creating users table...")
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			username VARCHAR(255) UNIQUE NOT NULL,
			password_hash VARCHAR(255) NOT NULL,
			role VARCHAR(50) NOT NULL
		)
	`)
	if err != nil {
		logger.Error("Failed to create users table: %v", err)
		return err
	}
	logger.Info("Users table created successfully")

	// Create questions table
	logger.Info("Creating questions table...")
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS questions (
			id SERIAL PRIMARY KEY,
			title VARCHAR(255) NOT NULL,
			statement TEXT NOT NULL,
			time_limit_ms INTEGER NOT NULL,
			memory_limit_mb INTEGER NOT NULL,
			status VARCHAR(50) NOT NULL,
			owner_id INTEGER NOT NULL,
			created_at TIMESTAMP NOT NULL,
			updated_at TIMESTAMP NOT NULL,
			test_input TEXT,
			test_output TEXT,
			FOREIGN KEY (owner_id) REFERENCES users(id)
		)
	`)
	if err != nil {
		logger.Error("Failed to create questions table: %v", err)
		return err
	}
	logger.Info("Questions table created successfully")

	return nil
}

func insertTestData(db *sql.DB) error {
	// Check if we already have test data
	logger.Info("Checking for existing test data...")
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM questions").Scan(&count)
	if err != nil {
		logger.Error("Failed to count questions: %v", err)
		return err
	}
	if count > 0 {
		logger.Info("Test data already exists (%d questions found), skipping insertion", count)
		return nil
	}

	// Insert test user if not exists
	logger.Info("Checking for test user...")
	var userID int
	err = db.QueryRow("SELECT id FROM users WHERE username = 'testuser'").Scan(&userID)
	if err == sql.ErrNoRows {
		logger.Info("Test user not found, creating...")
		_, err = db.Exec(`
			INSERT INTO users (username, password_hash, role)
			VALUES ('testuser', 'hashedpassword', 'user')
			RETURNING id
		`)
		if err != nil {
			logger.Error("Failed to create test user: %v", err)
			return err
		}
		err = db.QueryRow("SELECT id FROM users WHERE username = 'testuser'").Scan(&userID)
		if err != nil {
			logger.Error("Failed to get test user ID: %v", err)
			return err
		}
		logger.Info("Test user created with ID: %d", userID)
	} else {
		logger.Info("Test user found with ID: %d", userID)
	}

	// Insert test questions
	logger.Info("Inserting test questions...")
	now := time.Now().Format(time.RFC3339)
	_, err = db.Exec(`
		INSERT INTO questions (
			title, statement, time_limit_ms, memory_limit_mb,
			status, owner_id, created_at, updated_at,
			test_input, test_output
		) VALUES 
		('Test Published Question', 'This is a test published question', 1000, 128,
		 'published', $1, $2, $2, 'test input', 'test output'),
		('Test Draft Question', 'This is a test draft question', 1000, 128,
		 'draft', $1, $2, $2, 'test input', 'test output')
	`, userID, now)
	if err != nil {
		logger.Error("Failed to insert test questions: %v", err)
		return err
	}

	logger.Info("Successfully inserted test data")
	return nil
}
