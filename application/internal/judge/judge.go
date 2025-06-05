package judge

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
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
	TimeLimit int // seconds
	MemoryMB  int // megabytes
}

// Result represents the result of judging a submission
type Result struct {
	SubmissionID string
	Status       string
	Message      string
	TimeTaken    time.Duration
	MemoryUsed   int64
}

type runnerResult struct {
	Stdout   string `json:"stdout"`
	Stderr   string `json:"stderr"`
	ExitCode int    `json:"exit_code"`
	TimeMs   int64  `json:"time_ms"`
	Error    string `json:"error,omitempty"`
}

// Judge compiles and runs the submission against test cases
func Judge(submission Submission) (Result, error) {
	result := Result{SubmissionID: submission.ID}

	if strings.ToLower(submission.Language) != "go" {
		result.Status = "Unsupported Language"
		result.Message = "Only Go is supported."
		return result, nil
	}

	tempDir, err := os.MkdirTemp("", "judge-docker-*")
	if err != nil {
		return result, fmt.Errorf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	codePath := filepath.Join(tempDir, "code.go")
	inputPath := filepath.Join(tempDir, "input.txt")
	resultPath := filepath.Join(tempDir, "result.json")

	// Write code file
	if err := os.WriteFile(codePath, []byte(submission.Code), 0644); err != nil {
		return result, fmt.Errorf("failed to write code: %v", err)
	}

	totalStart := time.Now()
	for i, testCase := range submission.TestCases {
		if err := os.WriteFile(inputPath, []byte(testCase.Input), 0644); err != nil {
			result.Status = "Internal Error"
			result.Message = fmt.Sprintf("Failed to write input for test case %d: %v", i+1, err)
			return result, nil
		}

		// Build docker run command
		mem := fmt.Sprintf("%dm", submission.MemoryMB)
		cpu := "1"
		timeout := fmt.Sprintf("%d", submission.TimeLimit)
		dockerArgs := []string{
			"run", "--rm",
			"--cpus=" + cpu,
			"--memory=" + mem,
			"--network=none",
			"-v", tempDir + ":/code",
			"-e", "TIMEOUT_SEC=" + timeout,
			"code-runner-image:latest",
		}
		cmd := exec.Command("docker", dockerArgs...)
		var outBuf, errBuf bytes.Buffer
		cmd.Stdout = &outBuf
		cmd.Stderr = &errBuf
		if err := cmd.Run(); err != nil {
			result.Status = "Runtime Error"
			result.Message = fmt.Sprintf("Docker error (test case %d): %v, %s", i+1, err, errBuf.String())
			return result, nil
		}

		// Read result.json
		data, err := ioutil.ReadFile(resultPath)
		if err != nil {
			result.Status = "Internal Error"
			result.Message = fmt.Sprintf("Failed to read runner result (test case %d): %v", i+1, err)
			return result, nil
		}
		var rr runnerResult
		if err := json.Unmarshal(data, &rr); err != nil {
			result.Status = "Internal Error"
			result.Message = fmt.Sprintf("Failed to parse runner result (test case %d): %v", i+1, err)
			return result, nil
		}
		if rr.Error != "" {
			result.Status = "Runtime Error"
			result.Message = fmt.Sprintf("Test case %d: %s", i+1, rr.Error)
			return result, nil
		}
		if rr.ExitCode != 0 {
			result.Status = "Runtime Error"
			result.Message = fmt.Sprintf("Test case %d: %s", i+1, rr.Stderr)
			return result, nil
		}
		if !compareOutput(rr.Stdout, testCase.Output) {
			result.Status = "Wrong Answer"
			result.Message = fmt.Sprintf("Test case %d: Expected '%s', got '%s'", i+1, testCase.Output, rr.Stdout)
			return result, nil
		}
	}
	result.Status = "Accepted"
	result.TimeTaken = time.Since(totalStart)
	return result, nil
}

func compareOutput(actual, expected string) bool {
	return strings.TrimSpace(actual) == strings.TrimSpace(expected)
}
