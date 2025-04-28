package models

import "online-judge/internal/types"

// Submission represents a code submission from a user
type Submission struct {
	ID         int64            `json:"id"`
	QuestionID int64            `json:"question_id"`
	UserID     int64            `json:"user_id"`
	Code       string           `json:"code"`
	Language   string           `json:"language"`
	Status     string           `json:"status"`
	Message    string           `json:"message"`
	TimeTaken  int64            `json:"time_taken_ms"`
	MemoryUsed int64            `json:"memory_used_bytes"`
	CreatedAt  string           `json:"created_at"`
	TestCases  []types.TestCase `json:"test_cases"`
}

// TestCase represents a test case for a submission
type TestCase struct {
	Input  string `json:"input"`
	Output string `json:"output"`
}

// SubmissionResult represents the result of judging a submission
type SubmissionResult struct {
	SubmissionID string `json:"submission_id"`
	Status       string `json:"status"`
	Message      string `json:"message"`
	TimeTaken    int64  `json:"time_taken_ms"`
	MemoryUsed   int64  `json:"memory_used_bytes"`
}
