package models

import (
	"online-judge/internal/types"
	"strings"
)

type QuestionStatus string

const (
	StatusDraft     QuestionStatus = "draft"
	StatusPublished QuestionStatus = "published"
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
	TestInput     string         `json:"test_input"`
	TestOutput    string         `json:"test_output"`
}

// GetTestCases returns all test cases as pairs of input and output
func (q *Question) GetTestCases() []types.TestCase {
	if q.TestInput == "" || q.TestOutput == "" {
		return []types.TestCase{}
	}

	inputs := splitCSV(q.TestInput)
	outputs := splitCSV(q.TestOutput)

	// Use the minimum length to avoid index out of range
	numCases := min(len(inputs), len(outputs))
	cases := make([]types.TestCase, numCases)

	for i := 0; i < numCases; i++ {
		cases[i] = types.TestCase{
			Input:  inputs[i],
			Output: outputs[i],
		}
	}

	return cases
}

// SetTestCases sets the test cases from a slice of TestCase
func (q *Question) SetTestCases(cases []types.TestCase) {
	if len(cases) == 0 {
		q.TestInput = ""
		q.TestOutput = ""
		return
	}

	inputs := make([]string, len(cases))
	outputs := make([]string, len(cases))

	for i, tc := range cases {
		inputs[i] = tc.Input
		outputs[i] = tc.Output
	}

	q.TestInput = joinCSV(inputs)
	q.TestOutput = joinCSV(outputs)
}

// AddTestCase adds a single test case to the existing ones
func (q *Question) AddTestCase(input, output string) {
	cases := q.GetTestCases()
	cases = append(cases, types.TestCase{Input: input, Output: output})
	q.SetTestCases(cases)
}

// Helper functions
func splitCSV(s string) []string {
	if s == "" {
		return []string{}
	}
	return strings.Split(s, ",")
}

func joinCSV(parts []string) string {
	return strings.Join(parts, ",")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
