package repository

import (
	"database/sql"
	"errors"
	"online-judge/internal/models"
	"time"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) CreateUser(user *models.User) error {
	query := `
		INSERT INTO users (username, password_hash, role)
		VALUES ($1, $2, $3)
		RETURNING id
	`

	err := r.db.QueryRow(
		query,
		user.Username,
		user.Password,
		user.Role,
	).Scan(&user.ID)

	if err != nil {
		return err
	}

	return nil
}

func (r *UserRepository) GetUserByUsername(username string) (*models.User, error) {
	query := `
		SELECT id, username, password_hash, role
		FROM users
		WHERE username = $1
	`

	user := &models.User{}
	err := r.db.QueryRow(query, username).Scan(
		&user.ID,
		&user.Username,
		&user.Password,
		&user.Role,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return user, nil
}

func (r *UserRepository) UsernameExists(username string) (bool, error) {
	query := `
		SELECT EXISTS(SELECT 1 FROM users WHERE username = $1)
	`

	var exists bool
	err := r.db.QueryRow(query, username).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

func (r *UserRepository) EmailExists(email string) (bool, error) {
	query := `
		SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)
	`

	var exists bool
	err := r.db.QueryRow(query, email).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

func (r *UserRepository) GetUserBySession(sessionID string) (*models.User, error) {
	query := `
		SELECT u.id, u.username, u.password, u.role
		FROM users u
		JOIN sessions s ON u.id = s.user_id
		WHERE s.id = $1 AND s.expires_at > $2
	`

	user := &models.User{}
	err := r.db.QueryRow(query, sessionID, time.Now()).Scan(
		&user.ID,
		&user.Username,
		&user.Password,
		&user.Role,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return user, nil
}

func (r *UserRepository) UpdateUser(user *models.User) error {
	query := `
		UPDATE users 
		SET password_hash = $1, updated_at = $2
		WHERE username = $3
	`

	_, err := r.db.Exec(
		query,
		user.Password,
		time.Now(),
		user.Username,
	)

	if err != nil {
		return err
	}

	return nil
}

func (r *UserRepository) GetAllUsers() ([]*models.User, error) {
	query := `
		SELECT id, username, role
		FROM users
		ORDER BY username
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		user := &models.User{}
		err := rows.Scan(&user.ID, &user.Username, &user.Role)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func (r *UserRepository) UpdateUserRole(username string, newRole string) error {
	query := `
		UPDATE users 
		SET role = $1, updated_at = $2
		WHERE username = $3
	`

	_, err := r.db.Exec(
		query,
		newRole,
		time.Now(),
		username,
	)

	if err != nil {
		return err
	}

	return nil
}

func (r *UserRepository) IsLastAdmin(username string) (bool, error) {
	query := `
		SELECT COUNT(*) 
		FROM users 
		WHERE role = 'admin' AND username != $1
	`

	var count int
	err := r.db.QueryRow(query, username).Scan(&count)
	if err != nil {
		return false, err
	}

	return count == 0, nil
}
