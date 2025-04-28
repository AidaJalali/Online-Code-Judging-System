package types

// TestCase represents a test case for a question or submission
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

// Submission status constants
const (
	ResultOK           = "Ok"
	ResultCompileError = "Compile Error"
	ResultWrongAnswer  = "Wrong Answer"
	ResultMemoryLimit  = "Memory Limit"
	ResultTimeLimit    = "Time Limit"
	ResultRuntime      = "Runtime Error"
)
