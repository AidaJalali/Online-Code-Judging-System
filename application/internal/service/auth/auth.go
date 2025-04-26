package auth

import (
	"errors"
	"online-judge/internal/models"
	"online-judge/internal/repository"

	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepo repository.UserRepository
}

func NewAuthService(userRepo repository.UserRepository) *AuthService {
	return &AuthService{
		userRepo: userRepo,
	}
}

func (s *AuthService) Authenticate(username, password string) (*models.User, error) {
	user, err := s.userRepo.GetUserByUsername(username)
	if err != nil {
		return nil, err
	}

	if user.Password != password {
		return nil, nil
	}

	return user, nil
}

func (s *AuthService) Register(user *models.User) error {
	return s.userRepo.CreateUser(user)
}

func (s *AuthService) ChangePassword(userID int64, currentPassword, newPassword string) error {
	user, err := s.userRepo.GetUserByID(userID)
	if err != nil {
		return err
	}

	if user == nil {
		return errors.New("user not found")
	}

	// Verify current password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(currentPassword))
	if err != nil {
		return errors.New("current password is incorrect")
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// Update user password
	user.Password = string(hashedPassword)
	return s.userRepo.UpdateUser(user)
}
