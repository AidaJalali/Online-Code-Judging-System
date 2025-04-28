package judge

import (
	"bytes"
	"fmt"
	"online-judge/internal/types"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// Submission represents a code submission to be judged
type Submission struct {
	ID        string
	Code      string
	Language  string
	TestCases []types.TestCase
}

// Result represents the result of judging a submission
type Result struct {
	SubmissionID string
	Status       string
	Message      string
	TimeTaken    time.Duration
	MemoryUsed   int64
}

// Judge compiles and runs the submission against test cases
func Judge(submission Submission) (Result, error) {
	result := Result{
		SubmissionID: submission.ID,
	}

	// Create temporary directory for compilation
	tempDir, err := os.MkdirTemp("", "judge-*")
	if err != nil {
		return result, fmt.Errorf("failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Write submission code to file
	filePath := filepath.Join(tempDir, getFileName(submission.Language))
	if err := os.WriteFile(filePath, []byte(submission.Code), 0644); err != nil {
		return result, fmt.Errorf("failed to write submission file: %v", err)
	}

	// Compile the code
	if err := compileCode(filePath, submission.Language); err != nil {
		result.Status = "Compilation Error"
		result.Message = err.Error()
		return result, nil
	}

	// Run test cases
	startTime := time.Now()
	for i, testCase := range submission.TestCases {
		output, err := runTestCase(filePath, testCase.Input, submission.Language)
		if err != nil {
			result.Status = "Runtime Error"
			result.Message = fmt.Sprintf("Test case %d: %v", i+1, err)
			return result, nil
		}

		if !compareOutput(output, testCase.Output) {
			result.Status = "Wrong Answer"
			result.Message = fmt.Sprintf("Test case %d: Expected %s, got %s", i+1, testCase.Output, output)
			return result, nil
		}
	}

	result.Status = "Accepted"
	result.TimeTaken = time.Since(startTime)
	return result, nil
}

func getFileName(language string) string {
	switch strings.ToLower(language) {
	case "python":
		return "solution.py"
	case "java":
		return "Solution.java"
	case "cpp":
		return "solution.cpp"
	default:
		return "solution"
	}
}

func compileCode(filePath, language string) error {
	var cmd *exec.Cmd

	switch strings.ToLower(language) {
	case "python":
		// Python doesn't need compilation
		return nil
	case "java":
		cmd = exec.Command("javac", filePath)
	case "cpp":
		cmd = exec.Command("g++", "-std=c++17", "-O2", filePath, "-o", strings.TrimSuffix(filePath, ".cpp"))
	default:
		return fmt.Errorf("unsupported language: %s", language)
	}

	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("compilation error: %s", stderr.String())
	}

	return nil
}

func runTestCase(filePath, input, language string) (string, error) {
	var cmd *exec.Cmd

	switch strings.ToLower(language) {
	case "python":
		cmd = exec.Command("python", filePath)
	case "java":
		cmd = exec.Command("java", "-cp", filepath.Dir(filePath), "Solution")
	case "cpp":
		cmd = exec.Command(strings.TrimSuffix(filePath, ".cpp"))
	default:
		return "", fmt.Errorf("unsupported language: %s", language)
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Set timeout for execution
	cmd.Env = append(os.Environ(), "TIMEOUT=5")

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("runtime error: %s", stderr.String())
	}

	return strings.TrimSpace(stdout.String()), nil
}

func compareOutput(actual, expected string) bool {
	return strings.TrimSpace(actual) == strings.TrimSpace(expected)
}
