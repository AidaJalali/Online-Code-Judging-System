package repository

import (
	"context"
	"database/sql"
	"fmt"
	"online-judge/internal/models"
	"time"
)

// JudgeRepository defines the interface for judge-related database operations
type JudgeRepository interface {
	// GetSubmission retrieves a submission by ID
	GetSubmission(ctx context.Context, submissionID int64) (*models.Submission, error)

	// GetQuestion retrieves a question by ID
	GetQuestion(ctx context.Context, questionID int64) (*models.Question, error)

	// UpdateSubmission updates the submission status and results
	UpdateSubmission(ctx context.Context, submission *models.Submission) error
}

// judgeRepository implements the JudgeRepository interface
type judgeRepository struct {
	db *sql.DB
}

// NewJudgeRepository creates a new instance of JudgeRepository
func NewJudgeRepository(db *sql.DB) JudgeRepository {
	return &judgeRepository{db: db}
}

func (r *judgeRepository) GetSubmission(ctx context.Context, submissionID int64) (*models.Submission, error) {
	query := `
		SELECT id, user_id, question_id, code, language, status, error, time_used, memory_used, created_at, updated_at
		FROM submissions
		WHERE id = $1
	`

	submission := &models.Submission{}
	var timeUsedNanos int64
	err := r.db.QueryRowContext(ctx, query, submissionID).Scan(
		&submission.ID,
		&submission.UserID,
		&submission.QuestionID,
		&submission.Code,
		&submission.Language,
		&submission.Status,
		&submission.Error,
		&timeUsedNanos,
		&submission.MemoryUsed,
		&submission.CreatedAt,
		&submission.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get submission: %w", err)
	}

	submission.TimeUsed = time.Duration(timeUsedNanos)
	return submission, nil
}

func (r *judgeRepository) GetQuestion(ctx context.Context, questionID int64) (*models.Question, error) {
	query := `
		SELECT id, title, description, time_limit_ms, memory_limit_mb, status, owner_id, created_at, updated_at
		FROM questions
		WHERE id = $1
	`

	question := &models.Question{}
	err := r.db.QueryRowContext(ctx, query, questionID).Scan(
		&question.ID,
		&question.Title,
		&question.Description,
		&question.TimeLimitMs,
		&question.MemoryLimitMb,
		&question.Status,
		&question.OwnerID,
		&question.CreatedAt,
		&question.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get question: %w", err)
	}

	// Get test cases for the question
	testCasesQuery := `
		SELECT id, input, expected_output, is_public, created_at
		FROM test_cases
		WHERE question_id = $1
	`

	rows, err := r.db.QueryContext(ctx, testCasesQuery, questionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get test cases: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		testCase := models.TestCase{}
		err := rows.Scan(
			&testCase.ID,
			&testCase.Input,
			&testCase.ExpectedOutput,
			&testCase.IsPublic,
			&testCase.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan test case: %w", err)
		}
		testCase.QuestionID = questionID
		question.TestCases = append(question.TestCases, testCase)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating test cases: %w", err)
	}

	return question, nil
}

func (r *judgeRepository) UpdateSubmission(ctx context.Context, submission *models.Submission) error {
	query := `
		UPDATE submissions
		SET status = $1, error = $2, time_used = $3, memory_used = $4, updated_at = $5
		WHERE id = $6
	`

	_, err := r.db.ExecContext(ctx, query,
		submission.Status,
		submission.Error,
		int64(submission.TimeUsed),
		submission.MemoryUsed,
		time.Now(),
		submission.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update submission: %w", err)
	}

	return nil
}
