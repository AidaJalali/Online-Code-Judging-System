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
	query := `
		INSERT INTO questions (title, statement, time_limit_ms, memory_limit_mb, status, owner_id, test_input, test_output)
		VALUES ($1, $2, $3, $4, 'draft', $5, $6, $7)
		ON CONFLICT (owner_id, status) 
		WHERE status = 'draft'
		DO UPDATE SET 
			title = $1,
			statement = $2,
			time_limit_ms = $3,
			memory_limit_mb = $4,
			test_input = $6,
			test_output = $7
		RETURNING id
	`

	err := r.db.QueryRow(
		query,
		draft.Title,
		draft.Statement,
		draft.TimeLimitMs,
		draft.MemoryLimitMb,
		draft.OwnerID,
		draft.TestInput,
		draft.TestOutput,
	).Scan(&draft.ID)

	return err
}

func (r *DraftRepository) GetDraftByUserID(userID int64) (*models.Question, error) {
	query := `
		SELECT id, title, statement, time_limit_ms, memory_limit_mb, owner_id, test_input, test_output
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
