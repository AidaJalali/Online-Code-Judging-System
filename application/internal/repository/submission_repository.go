package repository

import (
	"database/sql"
	"online-judge/internal/judge"
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
	Language      string
	Status        string
	Message       string
	TimeTaken     int64
	MemoryUsed    int64
	CreatedAt     string
}

func (r *SubmissionRepository) GetUserSubmissionsWithQuestionTitle(userID int64) ([]SubmissionWithQuestion, error) {
	query := `
		SELECT s.id, s.question_id, q.title, s.language, s.status, s.message, s.time_taken, s.memory_used, s.created_at
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
		if err := rows.Scan(
			&s.ID,
			&s.QuestionID,
			&s.QuestionTitle,
			&s.Language,
			&s.Status,
			&s.Message,
			&s.TimeTaken,
			&s.MemoryUsed,
			&s.CreatedAt,
		); err != nil {
			return nil, err
		}
		submissions = append(submissions, s)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return submissions, nil
}

// SaveSubmissionResult saves the result of a code submission
func (r *SubmissionRepository) SaveSubmissionResult(result judge.Result) error {
	query := `
		UPDATE submissions
		SET status = $1,
			message = $2,
			time_taken = $3,
			memory_used = $4
		WHERE id = $5
	`
	_, err := r.db.Exec(query,
		result.Status,
		result.Message,
		result.TimeTaken.Milliseconds(),
		result.MemoryUsed,
		result.SubmissionID,
	)
	return err
}
