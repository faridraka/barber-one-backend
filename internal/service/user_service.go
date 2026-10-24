package service

import (
	"context"
	"fmt"

	"golang.org/x/crypto/bcrypt"

	"barber-one/backend/internal/repository"
)

// UserService handles user business logic
type UserService struct {
	repo *repository.Repository
}

func NewUserService(repo *repository.Repository) *UserService {
	return &UserService{repo: repo}
}

// Register creates a new user
func (s *UserService) Register(ctx context.Context, name, phone string, email *string, passwordHash string) (int64, error) {
	return s.repo.CreateUser(ctx, name, phone, email, passwordHash)
}

// Login validates user credentials and returns user data
func (s *UserService) Login(ctx context.Context, identifier, password string) (map[string]any, error) {
	user, err := s.repo.GetUserByIdentifier(ctx, identifier)
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	// Compare password hash
	passwordHash, ok := user["password_hash"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid credentials")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password)); err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	// Remove password from response
	delete(user, "password_hash")
	return user, nil
}

// GetProfile returns user profile by ID
func (s *UserService) GetProfile(ctx context.Context, userID int64) (map[string]any, error) {
	return s.repo.GetUserByID(ctx, userID)
}
