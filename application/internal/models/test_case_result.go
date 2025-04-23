package models

import "time"

// TestCaseResult represents the result of running a test case
type TestCaseResult struct {
	ID           int       `json:"id"`
	SubmissionID int       `json:"submission_id"`
	TestCaseID   int       `json:"test_case_id"`
	Status       string    `json:"status"` // "AC", "WA", "TLE", "MLE", "RE", "CE"
	Output       string    `json:"output"`
	Error        string    `json:"error"`
	TimeUsed     int       `json:"time_used"`   // in milliseconds
	MemoryUsed   int       `json:"memory_used"` // in kilobytes
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
