package repository

import (
	"database/sql"
	"online-judge/internal/errors"
	"online-judge/internal/logger"
	"online-judge/internal/models"
	"time"
)

type QuestionRepositoryImpl struct {
	db *sql.DB
}

func NewQuestionRepository(db *sql.DB) QuestionRepository {
	return &QuestionRepositoryImpl{db: db}
}

func (r *QuestionRepositoryImpl) CreateQuestion(question *models.Question) error {
	logger.Info("Creating question: %s", question.Title)

	tx, err := r.db.Begin()
	if err != nil {
		return errors.NewDatabaseError("Failed to start transaction", err)
	}
	defer tx.Rollback()

	// Insert question
	query := `
		INSERT INTO questions (
			title, description, time_limit_ms, memory_limit_mb,
			status, owner_id, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id`

	err = tx.QueryRow(
		query,
		question.Title,
		question.Description,
		question.TimeLimitMs,
		question.MemoryLimitMb,
		question.Status,
		question.OwnerID,
		question.CreatedAt,
		question.UpdatedAt,
	).Scan(&question.ID)

	if err != nil {
		return errors.NewDatabaseError("Failed to create question", err)
	}

	// Insert test cases
	for _, testCase := range question.TestCases {
		testCaseQuery := `
			INSERT INTO test_cases (
				question_id, input, expected_output, is_public,
				created_at, updated_at
			) VALUES ($1, $2, $3, $4, $5, $6)
			RETURNING id`

		err = tx.QueryRow(
			testCaseQuery,
			question.ID,
			testCase.Input,
			testCase.ExpectedOutput,
			testCase.IsPublic,
			time.Now(),
			time.Now(),
		).Scan(&testCase.ID)

		if err != nil {
			return errors.NewDatabaseError("Failed to create test case", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return errors.NewDatabaseError("Failed to commit transaction", err)
	}

	logger.Info("Successfully created question with ID %d", question.ID)
	return nil
}

func (r *QuestionRepositoryImpl) GetQuestion(id int) (*models.Question, error) {
	query := `
		SELECT id, title, description, time_limit_ms, memory_limit_mb,
			status, owner_id, created_at, updated_at
		FROM questions
		WHERE id = $1`

	question := &models.Question{}
	err := r.db.QueryRow(query, id).Scan(
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
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, errors.NewDatabaseError("Failed to get question", err)
	}

	// Get test cases
	testCasesQuery := `
		SELECT id, input, expected_output, is_public, created_at, updated_at
		FROM test_cases
		WHERE question_id = $1`

	rows, err := r.db.Query(testCasesQuery, id)
	if err != nil {
		return nil, errors.NewDatabaseError("Failed to get test cases", err)
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
			&testCase.UpdatedAt,
		)
		if err != nil {
			return nil, errors.NewDatabaseError("Failed to scan test case", err)
		}
		testCase.QuestionID = question.ID
		question.TestCases = append(question.TestCases, testCase)
	}

	return question, nil
}

func (r *QuestionRepositoryImpl) GetAllQuestions() ([]*models.Question, error) {
	query := `
		SELECT id, title, description, time_limit_ms, memory_limit_mb,
			status, owner_id, created_at, updated_at
		FROM questions
		ORDER BY created_at DESC`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, errors.NewDatabaseError("Failed to get questions", err)
	}
	defer rows.Close()

	var questions []*models.Question
	for rows.Next() {
		question := &models.Question{}
		err := rows.Scan(
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
			return nil, errors.NewDatabaseError("Failed to scan question", err)
		}

		// Get test cases for each question
		testCasesQuery := `
			SELECT id, input, expected_output, is_public, created_at, updated_at
			FROM test_cases
			WHERE question_id = $1`

		testCaseRows, err := r.db.Query(testCasesQuery, question.ID)
		if err != nil {
			return nil, errors.NewDatabaseError("Failed to get test cases", err)
		}
		defer testCaseRows.Close()

		for testCaseRows.Next() {
			testCase := models.TestCase{}
			err := testCaseRows.Scan(
				&testCase.ID,
				&testCase.Input,
				&testCase.ExpectedOutput,
				&testCase.IsPublic,
				&testCase.CreatedAt,
				&testCase.UpdatedAt,
			)
			if err != nil {
				return nil, errors.NewDatabaseError("Failed to scan test case", err)
			}
			testCase.QuestionID = question.ID
			question.TestCases = append(question.TestCases, testCase)
		}

		questions = append(questions, question)
	}

	return questions, nil
}

func (r *QuestionRepositoryImpl) UpdateQuestion(question *models.Question) error {
	tx, err := r.db.Begin()
	if err != nil {
		return errors.NewDatabaseError("Failed to start transaction", err)
	}
	defer tx.Rollback()

	// Update question
	query := `
		UPDATE questions
		SET title = $1, description = $2, time_limit_ms = $3,
			memory_limit_mb = $4, status = $5, updated_at = $6
		WHERE id = $7`

	_, err = tx.Exec(
		query,
		question.Title,
		question.Description,
		question.TimeLimitMs,
		question.MemoryLimitMb,
		question.Status,
		time.Now(),
		question.ID,
	)

	if err != nil {
		return errors.NewDatabaseError("Failed to update question", err)
	}

	// Delete existing test cases
	_, err = tx.Exec("DELETE FROM test_cases WHERE question_id = $1", question.ID)
	if err != nil {
		return errors.NewDatabaseError("Failed to delete test cases", err)
	}

	// Insert new test cases
	for _, testCase := range question.TestCases {
		testCaseQuery := `
			INSERT INTO test_cases (
				question_id, input, expected_output, is_public,
				created_at, updated_at
			) VALUES ($1, $2, $3, $4, $5, $6)
			RETURNING id`

		err = tx.QueryRow(
			testCaseQuery,
			question.ID,
			testCase.Input,
			testCase.ExpectedOutput,
			testCase.IsPublic,
			time.Now(),
			time.Now(),
		).Scan(&testCase.ID)

		if err != nil {
			return errors.NewDatabaseError("Failed to create test case", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return errors.NewDatabaseError("Failed to commit transaction", err)
	}

	return nil
}

func (r *QuestionRepositoryImpl) DeleteQuestion(id int) error {
	tx, err := r.db.Begin()
	if err != nil {
		return errors.NewDatabaseError("Failed to start transaction", err)
	}
	defer tx.Rollback()

	// Delete test cases first (due to foreign key constraint)
	_, err = tx.Exec("DELETE FROM test_cases WHERE question_id = $1", id)
	if err != nil {
		return errors.NewDatabaseError("Failed to delete test cases", err)
	}

	// Delete question
	_, err = tx.Exec("DELETE FROM questions WHERE id = $1", id)
	if err != nil {
		return errors.NewDatabaseError("Failed to delete question", err)
	}

	if err := tx.Commit(); err != nil {
		return errors.NewDatabaseError("Failed to commit transaction", err)
	}

	return nil
}
