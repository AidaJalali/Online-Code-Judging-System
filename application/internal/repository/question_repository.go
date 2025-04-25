package repository

import (
	"database/sql"
	"online-judge/internal/logger"
	"online-judge/internal/models"
)

type QuestionRepository struct {
	db *sql.DB
}

func NewQuestionRepository(db *sql.DB) *QuestionRepository {
	return &QuestionRepository{db: db}
}

func (r *QuestionRepository) CreateQuestion(question *models.Question) error {
	logger.Info("Creating question: %s", question.Title)

	query := `
		INSERT INTO questions (
			title, statement, time_limit_ms, memory_limit_mb,
			status, owner_id, created_at, updated_at,
			test_input, test_output
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id`

	err := r.db.QueryRow(
		query,
		question.Title,
		question.Statement,
		question.TimeLimitMs,
		question.MemoryLimitMb,
		question.Status,
		question.OwnerID,
		question.CreatedAt,
		question.UpdatedAt,
		question.TestInput,
		question.TestOutput,
	).Scan(&question.ID)

	if err != nil {
		logger.Error("Failed to create question in database: %v", err)
		return err
	}

	logger.Info("Successfully created question with ID %d", question.ID)
	return nil
}
