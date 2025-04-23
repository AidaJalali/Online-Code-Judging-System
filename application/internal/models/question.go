package models

type QuestionStatus string

const (
	StatusDraft QuestionStatus = "draft"
	// Add other status values as needed
)

type Question struct {
	ID            int64          `json:"id"`
	Title         string         `json:"title"`
	Statement     string         `json:"statement"`
	TimeLimitMs   int            `json:"time_limit_ms"`
	MemoryLimitMb int            `json:"memory_limit_mb"`
	Status        QuestionStatus `json:"status"`
	OwnerID       int64          `json:"owner_id"`
	CreatedAt     string         `json:"created_at"`
	UpdatedAt     string         `json:"updated_at"`
}

type TestCase struct {
	Input  string `json:"input"`
	Output string `json:"output"`
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
