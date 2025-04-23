package models

import "time"

// Submission represents a code submission
type Submission struct {
	ID         int
	UserID     int
	QuestionID int
	Code       string
	Language   string
	Status     string // "pending", "judging", "accepted", "wrong_answer", "time_limit_exceeded", "memory_limit_exceeded", "runtime_error", "compilation_error"
	Error      string // Error message if any
	TimeUsed   time.Duration
	MemoryUsed int64
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
