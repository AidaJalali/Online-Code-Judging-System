package models

// Submission represents a code submission in the system
type Submission struct {
	ID         int64  `json:"id"`
	QuestionID int64  `json:"question_id"`
	UserID     int64  `json:"user_id"`
	Code       string `json:"code"`
	Language   string `json:"language"`
	Status     string `json:"status"`
	CreatedAt  string `json:"created_at"`
}
