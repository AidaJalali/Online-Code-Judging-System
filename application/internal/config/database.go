package config

import (
	"database/sql"
	"fmt"
	"time"

	"online-judge/internal/logger"

	_ "github.com/lib/pq"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "mahdi"
	password = "secret123"
	dbname   = "online-judge"
)

func InitDB() (*sql.DB, error) {
	logger.Println("Initializing database connection...")
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	logger.Println("Opening database connection...")
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		logger.Println("Failed to open database: %v", err)
		return nil, fmt.Errorf("error opening database: %v", err)
	}

	logger.Println("Pinging database...")
	err = db.Ping()
	if err != nil {
		logger.Println("Failed to ping database: %v", err)
		return nil, fmt.Errorf("error connecting to database: %v", err)
	}
	logger.Println("Database connection successful")

	// Test query to verify access
	logger.Println("Testing database access...")
	var testResult int
	err = db.QueryRow("SELECT 1").Scan(&testResult)
	if err != nil {
		logger.Println("Failed to execute test query: %v", err)
		return nil, fmt.Errorf("error testing database access: %v", err)
	}
	logger.Println("Database access test successful")

	// Create tables if they don't exist
	logger.Println("Creating tables if they don't exist...")
	err = createTables(db)
	if err != nil {
		logger.Println("Failed to create tables: %v", err)
		return nil, fmt.Errorf("error creating tables: %v", err)
	}
	logger.Println("Tables created successfully")

	// Insert test data
	logger.Println("Inserting test data...")
	err = insertTestData(db)
	if err != nil {
		logger.Println("Error inserting test data: %v", err)
	} else {
		logger.Println("Test data inserted successfully")
	}

	// Verify test data
	logger.Println("Verifying test data...")
	var questionCount int
	err = db.QueryRow("SELECT COUNT(*) FROM questions").Scan(&questionCount)
	if err != nil {
		logger.Println("Failed to count questions: %v", err)
	} else {
		logger.Println("Found %d questions in database", questionCount)
	}

	var userCount int
	err = db.QueryRow("SELECT COUNT(*) FROM users").Scan(&userCount)
	if err != nil {
		logger.Println("Failed to count users: %v", err)
	} else {
		logger.Println("Found %d users in database", userCount)
	}

	return db, nil
}

func createTables(db *sql.DB) error {
	// Create users table
	logger.Println("Creating users table...")
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			username VARCHAR(255) UNIQUE NOT NULL,
			password_hash VARCHAR(255) NOT NULL,
			role VARCHAR(50) NOT NULL
		)
	`)
	if err != nil {
		logger.Println("Failed to create users table: %v", err)
		return err
	}
	logger.Println("Users table created successfully")

	// Create questions table
	logger.Println("Creating questions table...")
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
			test_input varchar(50),
			test_output varchar(50),
			FOREIGN KEY (owner_id) REFERENCES users(id)
		)
	`)
	if err != nil {
		logger.Println("Failed to create questions table: %v", err)
		return err
	}
	logger.Println("Questions table created successfully")

	// Create submissions table
	logger.Println("Creating submissions table...")
	_, err = db.Exec(`
		DROP TABLE IF EXISTS submissions;
		DROP TYPE IF EXISTS submission_status;

		CREATE TYPE submission_status AS ENUM (
			'Accepted',
			'Wrong Answer',
			'Compilation Error',
			'Runtime Error',
			'Time Limit Exceeded',
			'Memory Limit Exceeded',
			'Pending'
		);

		CREATE TABLE submissions (
			id SERIAL PRIMARY KEY,
			question_id INTEGER NOT NULL,
			user_id INTEGER NOT NULL,
			code TEXT NOT NULL,
			language VARCHAR(50) NOT NULL,
			status submission_status NOT NULL DEFAULT 'Pending',
			message TEXT,
			time_taken BIGINT,
			memory_used BIGINT,
			created_at TIMESTAMP NOT NULL,
			FOREIGN KEY (question_id) REFERENCES questions(id),
			FOREIGN KEY (user_id) REFERENCES users(id)
		)
	`)
	if err != nil {
		logger.Println("Failed to create submissions table: %v", err)
		return err
	}
	logger.Println("Submissions table created successfully")

	return nil
}

func insertTestData(db *sql.DB) error {
	// Check if we already have test data
	logger.Println("Checking for existing test data...")
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM questions").Scan(&count)
	if err != nil {
		logger.Println("Failed to count questions: %v", err)
		return err
	}
	if count > 0 {
		logger.Println("Test data already exists (%d questions found), skipping insertion", count)
		return nil
	}

	// Insert test user if not exists
	logger.Println("Checking for test user...")
	var userID int
	err = db.QueryRow("SELECT id FROM users WHERE username = 'testuser'").Scan(&userID)
	if err == sql.ErrNoRows {
		logger.Println("Test user not found, creating...")
		_, err = db.Exec(`
			INSERT INTO users (username, password_hash, role)
			VALUES ('testuser', 'hashedpassword', 'user')
			RETURNING id
		`)
		if err != nil {
			logger.Println("Failed to create test user: %v", err)
			return err
		}
		err = db.QueryRow("SELECT id FROM users WHERE username = 'testuser'").Scan(&userID)
		if err != nil {
			logger.Println("Failed to get test user ID: %v", err)
			return err
		}
		logger.Println("Test user created with ID: %d", userID)
	} else {
		logger.Println("Test user found with ID: %d", userID)
	}

	// Insert test questions
	logger.Println("Inserting test questions...")
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
		logger.Println("Failed to insert test questions: %v", err)
		return err
	}

	logger.Println("Successfully inserted test data")
	return nil
}
