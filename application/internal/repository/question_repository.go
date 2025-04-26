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

func (r *QuestionRepository) GetAllQuestions() ([]models.Question, error) {
	query := `
		SELECT id, title, statement, time_limit_ms, memory_limit_mb, 
		       status, owner_id, created_at, updated_at, test_input, test_output
		FROM questions
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var questions []models.Question
	for rows.Next() {
		var q models.Question
		err := rows.Scan(
			&q.ID,
			&q.Title,
			&q.Statement,
			&q.TimeLimitMs,
			&q.MemoryLimitMb,
			&q.Status,
			&q.OwnerID,
			&q.CreatedAt,
			&q.UpdatedAt,
			&q.TestInput,
			&q.TestOutput,
		)
		if err != nil {
			return nil, err
		}
		questions = append(questions, q)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return questions, nil
}
