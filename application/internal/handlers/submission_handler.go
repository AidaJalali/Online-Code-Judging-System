package handlers

import (
	"encoding/json"
	"net/http"

	"online-judge/internal/judge"
	"online-judge/internal/models"
	"online-judge/internal/types"
)

// SubmissionHandler handles code submission requests
type SubmissionHandler struct {
	// Add any dependencies here (e.g., database connection)
}

// NewSubmissionHandler creates a new submission handler
func NewSubmissionHandler() *SubmissionHandler {
	return &SubmissionHandler{}
}

// HandleSubmission processes a code submission
func (h *SubmissionHandler) HandleSubmission(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var submission models.Submission
	if err := json.NewDecoder(r.Body).Decode(&submission); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Convert to judge.Submission
	judgeSubmission := judge.Submission{
		ID:        string(submission.ID),
		Code:      submission.Code,
		Language:  submission.Language,
		TestCases: make([]types.TestCase, len(submission.TestCases)),
	}

	for i, tc := range submission.TestCases {
		judgeSubmission.TestCases[i] = types.TestCase{
			Input:  tc.Input,
			Output: tc.Output,
		}
	}

	// Judge the submission
	result, err := judge.Judge(judgeSubmission)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Convert result to response
	response := types.SubmissionResult{
		SubmissionID: result.SubmissionID,
		Status:       result.Status,
		Message:      result.Message,
		TimeTaken:    result.TimeTaken.Milliseconds(),
		MemoryUsed:   result.MemoryUsed,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
