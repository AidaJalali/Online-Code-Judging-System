package repository

import (
	"database/sql"
	"online-judge/internal/models"
)

type DraftRepository struct {
	db *sql.DB
}

func NewDraftRepository(db *sql.DB) *DraftRepository {
	return &DraftRepository{db: db}
}

func (r *DraftRepository) SaveDraft(draft *models.Question) error {
	// First, try to get existing draft
	existingDraft, err := r.GetDraftByUserID(draft.OwnerID)
	if err != nil && err != sql.ErrNoRows {
		return err
	}

	if existingDraft != nil {
		// Update existing draft
		query := `
			UPDATE questions 
			SET title = $1,
				statement = $2,
				time_limit_ms = $3,
				memory_limit_mb = $4,
				updated_at = NOW(),
				test_input = $5,
				test_output = $6
			WHERE owner_id = $7 AND status = 'draft'
			RETURNING id
		`

		err = r.db.QueryRow(
			query,
			draft.Title,
			draft.Statement,
			draft.TimeLimitMs,
			draft.MemoryLimitMb,
			draft.TestInput,
			draft.TestOutput,
			draft.OwnerID,
		).Scan(&draft.ID)
	} else {
		// Insert new draft
		query := `
			INSERT INTO questions (
				title, statement, time_limit_ms, memory_limit_mb, 
				status, created_at, updated_at, owner_id, 
				test_input, test_output
			) VALUES ($1, $2, $3, $4, 'draft', NOW(), NOW(), $5, $6, $7)
			RETURNING id
		`

		err = r.db.QueryRow(
			query,
			draft.Title,
			draft.Statement,
			draft.TimeLimitMs,
			draft.MemoryLimitMb,
			draft.OwnerID,
			draft.TestInput,
			draft.TestOutput,
		).Scan(&draft.ID)
	}

	return err
}

func (r *DraftRepository) GetDraftByUserID(userID int64) (*models.Question, error) {
	query := `
		SELECT id, title, statement, time_limit_ms, memory_limit_mb, owner_id, created_at, updated_at, test_input, test_output
		FROM questions
		WHERE owner_id = $1 AND status = 'draft'
	`

	draft := &models.Question{OwnerID: userID}
	err := r.db.QueryRow(query, userID).Scan(
		&draft.ID,
		&draft.Title,
		&draft.Statement,
		&draft.TimeLimitMs,
		&draft.MemoryLimitMb,
		&draft.OwnerID,
		&draft.CreatedAt,
		&draft.UpdatedAt,
		&draft.TestInput,
		&draft.TestOutput,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	draft.Status = models.StatusDraft
	return draft, nil
}

func (r *DraftRepository) DeleteDraft(userID int64) error {
	query := `DELETE FROM questions WHERE owner_id = $1 AND status = 'draft'`
	_, err := r.db.Exec(query, userID)
	return err
}

func (r *DraftRepository) GetDraftsByUserID(userID int64) ([]models.Question, error) {
	query := `
		SELECT id, title, statement, time_limit_ms, memory_limit_mb, 
		       status, owner_id, created_at, updated_at, test_input, test_output
		FROM questions
		WHERE owner_id = $1 AND status = 'draft'
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var drafts []models.Question
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
		drafts = append(drafts, q)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return drafts, nil
}
