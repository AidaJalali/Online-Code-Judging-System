package models

import "time"

type QuestionStatus string

const (
	StatusDraft QuestionStatus = "draft"
	// Add other status values as needed
)

// Question represents a coding question
type Question struct {
	ID            int64          `json:"id"`
	Title         string         `json:"title"`
	Description   string         `json:"description"`
	Difficulty    string         `json:"difficulty"` // "easy", "medium", "hard"
	TimeLimitMs   int            `json:"time_limit_ms"`
	MemoryLimitMb int            `json:"memory_limit_mb"`
	TestCases     []TestCase     `json:"test_cases"`
	Status        QuestionStatus `json:"status"`
	OwnerID       int64          `json:"owner_id"`
	CreatedBy     int64          `json:"created_by"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
}

// TestCase represents a test case for a question
type TestCase struct {
	ID             int64     `json:"id"`
	QuestionID     int64     `json:"question_id"`
	Input          string    `json:"input"`
	ExpectedOutput string    `json:"expected_output"`
	IsPublic       bool      `json:"is_public"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type SubmissionResult string

const (
	ResultOK           SubmissionResult = "Ok"
	ResultCompileError SubmissionResult = "Compile Error"
	ResultWrongAnswer  SubmissionResult = "Wrong Answer"
	ResultMemoryLimit  SubmissionResult = "Memory Limit"
	ResultTimeLimit    SubmissionResult = "Time Limit"
	ResultRuntime      SubmissionResult = "Runtime Error"
)
