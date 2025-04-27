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
		SELECT id, question_id, user_id, code, status, created_at
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

func (r *SubmissionRepository) CreateSubmission(sub *models.Submission) error {
	query := `
		INSERT INTO submissions (question_id, user_id, code, status, created_at)
		VALUES ($1, $2, $3, $4, NOW())
		RETURNING id, created_at
	`
	return r.db.QueryRow(query, sub.QuestionID, sub.UserID, sub.Code, sub.Status).Scan(&sub.ID, &sub.CreatedAt)
}

type SubmissionWithQuestion struct {
	ID            int64
	QuestionID    int64
	QuestionTitle string
	Status        string
	CreatedAt     string
}

func (r *SubmissionRepository) GetUserSubmissionsWithQuestionTitle(userID int64) ([]SubmissionWithQuestion, error) {
	query := `
		SELECT s.id, s.question_id, q.title, s.status, s.created_at
		FROM submissions s
		JOIN questions q ON s.question_id = q.id
		WHERE s.user_id = $1
		ORDER BY s.created_at DESC
	`
	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var submissions []SubmissionWithQuestion
	for rows.Next() {
		var s SubmissionWithQuestion
		if err := rows.Scan(&s.ID, &s.QuestionID, &s.QuestionTitle, &s.Status, &s.CreatedAt); err != nil {
			return nil, err
		}
		submissions = append(submissions, s)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return submissions, nil
}
