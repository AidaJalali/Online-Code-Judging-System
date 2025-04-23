package repository

import (
	"database/sql"
	"online-judge/internal/logger"
	"online-judge/internal/models"
)

type Repository struct {
	db *sql.DB
}

func (r *Repository) CreateUser(user *models.User) error {
	logger.Info("Creating user in database with role: %s", user.Role)
	_, err := r.db.Exec(`
		INSERT INTO users (username, password, role)
		VALUES ($1, $2, $3)
	`, user.Username, user.Password, user.Role)
	if err != nil {
		logger.Error("Failed to create user in database: %v", err)
		return err
	}
	logger.Info("Successfully created user %s with role %s in database", user.Username, user.Role)
	return nil
}
