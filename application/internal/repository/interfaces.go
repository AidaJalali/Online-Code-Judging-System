package repository

import (
	"context"
	"online-judge/internal/models"
)

type UserRepository interface {
	CreateUser(user *models.User) error
	GetUserByUsername(username string) (*models.User, error)
	GetUserByID(id int64) (*models.User, error)
	GetUserBySession(sessionID string) (*models.User, error)
	UsernameExists(username string) (bool, error)
	UpdateUser(user *models.User) error
	GetAllUsers() ([]*models.User, error)
	UpdateUserRole(username string, newRole string) error
	IsLastAdmin(username string) (bool, error)
}

type QuestionRepository interface {
	CreateQuestion(question *models.Question) error
	GetQuestion(id int) (*models.Question, error)
	GetAllQuestions() ([]*models.Question, error)
	UpdateQuestion(question *models.Question) error
	DeleteQuestion(id int) error
}

type SubmissionRepository interface {
	CreateSubmission(submission *models.Submission) error
	GetSubmission(ctx context.Context, submissionID int64) (*models.Submission, error)
	GetSubmissionsByUserID(userID int) ([]*models.Submission, error)
	UpdateSubmission(ctx context.Context, submission *models.Submission) error
}
