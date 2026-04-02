package service

import (
	"errors"
	"fmt"

	"github.com/username/kafe-backend/internal/models"
	"github.com/username/kafe-backend/internal/pkg/security"
	"github.com/username/kafe-backend/internal/repository"
)

type AuthService struct {
	userRepo *repository.UserRepository
}

func NewAuthService(userRepo *repository.UserRepository) *AuthService {
	return &AuthService{userRepo: userRepo}
}

func (s *AuthService) Register(fullName, phone, password string, role models.UserRole) (*models.User, string, error) {
	// Check if already exists
	existing, err := s.userRepo.GetByPhone(phone)
	if err != nil {
		return nil, "", err
	}
	if existing != nil {
		return nil, "", errors.New("user already exists")
	}

	hash, err := security.HashPassword(password)
	if err != nil {
		return nil, "", fmt.Errorf("failed to hash password: %w", err)
	}

	// First user logic: If no users exist, make this one an admin
	count, _ := s.userRepo.Count()
	if count == 0 {
		role = models.RoleAdmin
	}

	user := &models.User{
		FullName:     fullName,
		Phone:        phone,
		PasswordHash: hash,
		Role:         role,
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, "", err
	}

	token, err := security.GenerateToken(user.ID, string(user.Role))
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate token: %w", err)
	}

	return user, token, nil
}

func (s *AuthService) Login(phone, password string) (*models.User, string, error) {
	user, err := s.userRepo.GetByPhone(phone)
	if err != nil {
		return nil, "", err
	}
	if user == nil {
		return nil, "", errors.New("invalid phone or password")
	}

	if !security.CheckPasswordHash(password, user.PasswordHash) {
		return nil, "", errors.New("invalid phone or password")
	}

	token, err := security.GenerateToken(user.ID, string(user.Role))
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate token: %w", err)
	}

	return user, token, nil
}
