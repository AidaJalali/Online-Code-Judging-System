package service

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"online-judge/internal/repository"
)

type JudgeService struct {
	judgeRepo repository.JudgeRepository
}

func NewJudgeService(judgeRepo repository.JudgeRepository) *JudgeService {
	return &JudgeService{
		judgeRepo: judgeRepo,
	}
}

func (s *JudgeService) JudgeSubmission(ctx context.Context, submissionID int64) error {
	// Get submission details
	submission, err := s.judgeRepo.GetSubmission(ctx, submissionID)
	if err != nil {
		return fmt.Errorf("failed to get submission: %w", err)
	}

	// Get question details
	question, err := s.judgeRepo.GetQuestion(ctx, int64(submission.QuestionID))
	if err != nil {
		return fmt.Errorf("failed to get question: %w", err)
	}

	// Create temporary directory for compilation
	tempDir, err := os.MkdirTemp("", "judge-*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Write submission code to file
	codeFile := filepath.Join(tempDir, "solution.go")
	if err := os.WriteFile(codeFile, []byte(submission.Code), 0644); err != nil {
		return fmt.Errorf("failed to write code file: %w", err)
	}

	// Compile the code
	cmd := exec.Command("go", "build", "-o", filepath.Join(tempDir, "solution"), codeFile)
	if output, err := cmd.CombinedOutput(); err != nil {
		// Update submission status with compilation error
		submission.Status = "compilation_error"
		submission.Error = string(output)
		if err := s.judgeRepo.UpdateSubmission(ctx, submission); err != nil {
			return fmt.Errorf("failed to update submission: %w", err)
		}
		return nil
	}

	// Run test cases
	for i, testCase := range question.TestCases {
		// Execute the program with test case input
		cmd := exec.Command(filepath.Join(tempDir, "solution"))
		cmd.Stdin = strings.NewReader(testCase.Input)
		output, err := cmd.CombinedOutput()
		if err != nil {
			// Update submission status with runtime error
			submission.Status = "runtime_error"
			submission.Error = fmt.Sprintf("Test case %d: %v", i+1, err)
			if err := s.judgeRepo.UpdateSubmission(ctx, submission); err != nil {
				return fmt.Errorf("failed to update submission: %w", err)
			}
			return nil
		}

		// Compare output with expected output
		if strings.TrimSpace(string(output)) != strings.TrimSpace(testCase.ExpectedOutput) {
			// Update submission status with wrong answer
			submission.Status = "wrong_answer"
			submission.Error = fmt.Sprintf("Test case %d: Output does not match expected output", i+1)
			if err := s.judgeRepo.UpdateSubmission(ctx, submission); err != nil {
				return fmt.Errorf("failed to update submission: %w", err)
			}
			return nil
		}
	}

	// All test cases passed
	submission.Status = "accepted"
	if err := s.judgeRepo.UpdateSubmission(ctx, submission); err != nil {
		return fmt.Errorf("failed to update submission: %w", err)
	}

	return nil
}
