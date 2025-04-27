package repository

import (
	"database/sql"
	"online-judge/internal/models"
)

type SubmissionRepository struct {
	db *sql.DB
}

func NewSubmissionRepository(db *sql.DB) *SubmissionRepository {
	return &SubmissionRepository{db: db}
}

func (r *SubmissionRepository) GetAllSubmissions() ([]models.Submission, error) {
	query := `
		SELECT id, question_id, user_id, code, language, status, created_at
		FROM submissions
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var submissions []models.Submission
	for rows.Next() {
		var s models.Submission
		err := rows.Scan(
			&s.ID,
			&s.QuestionID,
			&s.UserID,
			&s.Code,
			&s.Language,
			&s.Status,
			&s.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		submissions = append(submissions, s)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return submissions, nil
}
