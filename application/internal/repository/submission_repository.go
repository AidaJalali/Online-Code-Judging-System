package repository

import (
	"context"
	"database/sql"
	"online-judge/internal/models"
	"time"
)

type SubmissionRepositoryImpl struct {
	db *sql.DB
}

func NewSubmissionRepository(db *sql.DB) *SubmissionRepositoryImpl {
	return &SubmissionRepositoryImpl{db: db}
}

func (r *SubmissionRepositoryImpl) CreateSubmission(submission *models.Submission) error {
	query := `INSERT INTO submissions (user_id, question_id, code, language, status, error, time_used, memory_used, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	now := time.Now()
	_, err := r.db.Exec(query,
		submission.UserID,
		submission.QuestionID,
		submission.Code,
		submission.Language,
		submission.Status,
		submission.Error,
		submission.TimeUsed,
		submission.MemoryUsed,
		now,
		now,
	)
	return err
}

func (r *SubmissionRepositoryImpl) GetSubmission(ctx context.Context, submissionID int64) (*models.Submission, error) {
	query := `SELECT id, user_id, question_id, code, language, status, error, time_used, memory_used, created_at, updated_at
		FROM submissions WHERE id = ?`

	submission := &models.Submission{}
	err := r.db.QueryRowContext(ctx, query, submissionID).Scan(
		&submission.ID,
		&submission.UserID,
		&submission.QuestionID,
		&submission.Code,
		&submission.Language,
		&submission.Status,
		&submission.Error,
		&submission.TimeUsed,
		&submission.MemoryUsed,
		&submission.CreatedAt,
		&submission.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return submission, nil
}

func (r *SubmissionRepositoryImpl) GetSubmissionsByUserID(userID int) ([]*models.Submission, error) {
	query := `SELECT id, user_id, question_id, code, language, status, error, time_used, memory_used, created_at, updated_at
		FROM submissions WHERE user_id = ? ORDER BY created_at DESC`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var submissions []*models.Submission
	for rows.Next() {
		submission := &models.Submission{}
		err := rows.Scan(
			&submission.ID,
			&submission.UserID,
			&submission.QuestionID,
			&submission.Code,
			&submission.Language,
			&submission.Status,
			&submission.Error,
			&submission.TimeUsed,
			&submission.MemoryUsed,
			&submission.CreatedAt,
			&submission.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		submissions = append(submissions, submission)
	}
	return submissions, nil
}

func (r *SubmissionRepositoryImpl) UpdateSubmission(ctx context.Context, submission *models.Submission) error {
	query := `UPDATE submissions SET status = ?, error = ?, time_used = ?, memory_used = ?, updated_at = ?
		WHERE id = ?`

	now := time.Now()
	_, err := r.db.ExecContext(ctx, query,
		submission.Status,
		submission.Error,
		submission.TimeUsed,
		submission.MemoryUsed,
		now,
		submission.ID,
	)
	return err
}
