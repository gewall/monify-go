package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/alginugraha/monify/internal/domain"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	users UserRepo
}

func NewAuthService(users UserRepo) *AuthService {
	return &AuthService{users: users}
}

// Authenticate verifies credentials and returns the user on success.
// It returns domain.ErrUnauthorized for any bad-credentials case.
func (s *AuthService) Authenticate(ctx context.Context, email, password string) (domain.User, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	u, err := s.users.ByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return domain.User{}, domain.ErrUnauthorized
		}
		return domain.User{}, err
	}
	if bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)) != nil {
		return domain.User{}, domain.ErrUnauthorized
	}
	return u, nil
}

// CreateUser registers a new user with a bcrypt-hashed password.
func (s *AuthService) CreateUser(ctx context.Context, email, password string) (domain.User, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	if email == "" || len(password) < 8 {
		return domain.User{}, fmt.Errorf("%w: email required and password must be >= 8 chars", domain.ErrInvalid)
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return domain.User{}, err
	}
	return s.users.Create(ctx, email, string(hash))
}

// ByID looks up a user, used by auth middleware to confirm the session subject.
func (s *AuthService) ByID(ctx context.Context, id string) (domain.User, error) {
	return s.users.ByID(ctx, id)
}
