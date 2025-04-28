package repository

import (
	"database/sql"
	"fmt"
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

func (r *QuestionRepository) GetQuestionByID(id string) (*models.Question, error) {
	logger.Info("Fetching question with ID: %s", id)

	query := `
		SELECT id, title, statement, time_limit_ms, memory_limit_mb, 
		       status, owner_id, created_at, updated_at, test_input, test_output
		FROM questions
		WHERE id = $1
	`

	var question models.Question
	err := r.db.QueryRow(query, id).Scan(
		&question.ID,
		&question.Title,
		&question.Statement,
		&question.TimeLimitMs,
		&question.MemoryLimitMb,
		&question.Status,
		&question.OwnerID,
		&question.CreatedAt,
		&question.UpdatedAt,
		&question.TestInput,
		&question.TestOutput,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			logger.Info("Question with ID %s not found", id)
			return nil, nil
		}
		logger.Error("Failed to fetch question from database: %v", err)
		return nil, err
	}

	logger.Info("Successfully fetched question with ID %s", id)
	return &question, nil
}

func (r *QuestionRepository) GetPublishedQuestions() ([]models.Question, error) {
	logger.Info("Getting all published questions")

	// First check if we can connect to the database
	err := r.db.Ping()
	if err != nil {
		logger.Error("Database connection failed: %v", err)
		return nil, err
	}
	logger.Info("Database connection successful")

	query := `
		SELECT id, title, statement, time_limit_ms, memory_limit_mb, 
			   status, owner_id, created_at, updated_at, test_input, test_output
		FROM questions
		WHERE status = 'published'
		ORDER BY created_at DESC
	`
	logger.Info("Executing query: %s", query)

	rows, err := r.db.Query(query)
	if err != nil {
		logger.Error("Failed to query published questions: %v", err)
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
			logger.Error("Failed to scan published question: %v", err)
			return nil, err
		}
		logger.Info("Found question: ID=%d, Title=%s, Status=%s", q.ID, q.Title, q.Status)
		questions = append(questions, q)
	}

	if err = rows.Err(); err != nil {
		logger.Error("Error iterating published questions: %v", err)
		return nil, err
	}

	logger.Info("Successfully retrieved %d published questions", len(questions))
	return questions, nil
}

func (r *QuestionRepository) GetDraftQuestionsByUser(userID int64) ([]models.Question, error) {
	query := `
		SELECT id, title, statement, time_limit_ms, memory_limit_mb, 
		       status, owner_id, created_at, updated_at, test_input, test_output
		FROM questions
		WHERE status = 'draft' AND owner_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query, userID)
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

func (r *QuestionRepository) UpdateQuestion(question *models.Question) error {
	query := `
		UPDATE questions 
		SET title = $1, 
		    statement = $2, 
		    time_limit_ms = $3,
		    memory_limit_mb = $4,
		    status = $5, 
		    test_input = $6,
		    test_output = $7,
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = $8
	`
	_, err := r.db.Exec(query,
		question.Title,
		question.Statement,
		question.TimeLimitMs,
		question.MemoryLimitMb,
		question.Status,
		question.TestInput,
		question.TestOutput,
		question.ID)
	if err != nil {
		return fmt.Errorf("failed to update question: %w", err)
	}
	return nil
}

func (r *QuestionRepository) DeleteQuestion(id string) error {
	query := `DELETE FROM questions WHERE id = $1`
	_, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete question: %w", err)
	}
	return nil
}

// GetDraftQuestions returns all draft questions
func (r *QuestionRepository) GetDraftQuestions() ([]models.Question, error) {
	query := `
		SELECT id, title, statement, time_limit_ms, memory_limit_mb, 
			   status, owner_id, created_at, updated_at, test_input, test_output
		FROM questions
		WHERE status = 'draft'
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

// GetDraftsByUserID returns all draft questions for a specific user
func (r *QuestionRepository) GetDraftsByUserID(userID int64) ([]models.Question, error) {
	logger.Info("Getting draft questions for user ID: %d", userID)

	// First check if we can connect to the database
	err := r.db.Ping()
	if err != nil {
		logger.Error("Database connection failed: %v", err)
		return nil, err
	}
	logger.Info("Database connection successful")

	query := `
		SELECT id, title, statement, time_limit_ms, memory_limit_mb, 
			   status, owner_id, created_at, updated_at, test_input, test_output
		FROM questions
		WHERE status = 'draft' AND owner_id = $1
		ORDER BY created_at DESC
	`
	logger.Info("Executing query: %s with userID=%d", query, userID)

	rows, err := r.db.Query(query, userID)
	if err != nil {
		logger.Error("Failed to query draft questions for user %d: %v", userID, err)
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
			logger.Error("Failed to scan draft question for user %d: %v", userID, err)
			return nil, err
		}
		logger.Info("Found draft question: ID=%d, Title=%s, Status=%s", q.ID, q.Title, q.Status)
		questions = append(questions, q)
	}

	if err = rows.Err(); err != nil {
		logger.Error("Error iterating draft questions for user %d: %v", userID, err)
		return nil, err
	}

	logger.Info("Successfully retrieved %d draft questions for user %d", len(questions), userID)
	return questions, nil
}
