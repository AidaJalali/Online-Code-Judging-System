package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"online-judge/internal/models"
	"time"

	"golang.org/x/crypto/bcrypt"
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

func (r *UserRepository) UpdateUserRole(username, newRole string) error {
	// First verify the user exists
	user, err := r.GetUserByUsername(username)
	if err != nil {
		return fmt.Errorf("error getting user: %v", err)
	}
	if user == nil {
		return fmt.Errorf("user not found: %s", username)
	}

	// Convert role to proper enum value
	var roleValue string
	switch newRole {
	case "admin":
		roleValue = "admin"
	case "user":
		roleValue = "regular" // The enum value is 'regular' instead of 'user'
	default:
		return fmt.Errorf("invalid role: %s", newRole)
	}

	// Update the role
	query := "UPDATE users SET role = $1 WHERE username = $2"
	result, err := r.db.Exec(query, roleValue, username)
	if err != nil {
		return fmt.Errorf("error executing update: %v", err)
	}

	// Check if any rows were affected
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %v", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("no rows were updated for user %s", username)
	}

	return nil
}

func (r *UserRepository) GetAllUsers(page, pageSize int) ([]*models.User, int, error) {
	// Calculate offset
	offset := (page - 1) * pageSize

	// Get total count
	var total int
	err := r.db.QueryRow("SELECT COUNT(*) FROM users").Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("error getting total users: %v", err)
	}

	// Get paginated users
	query := "SELECT id, username, role FROM users ORDER BY username LIMIT $1 OFFSET $2"
	rows, err := r.db.Query(query, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("error querying users: %v", err)
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		user := &models.User{}
		if err := rows.Scan(&user.ID, &user.Username, &user.Role); err != nil {
			return nil, 0, fmt.Errorf("error scanning user: %v", err)
		}
		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating users: %v", err)
	}

	return users, total, nil
}

func (r *UserRepository) VerifyPassword(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}
