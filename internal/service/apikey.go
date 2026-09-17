package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/alginugraha/monify/internal/domain"
	"github.com/alginugraha/monify/internal/platform/apikey"
)

type APIKeyService struct {
	repo  APIKeyRepo
	users UserRepo
}

func NewAPIKeyService(repo APIKeyRepo, users UserRepo) *APIKeyService {
	return &APIKeyService{repo: repo, users: users}
}

// Create mints a new API key for a user. The plaintext token is returned
// only here; callers must save it, since it cannot be recovered afterwards.
func (s *APIKeyService) Create(ctx context.Context, userID, name string) (domain.APIKey, string, error) {
	if strings.TrimSpace(name) == "" {
		return domain.APIKey{}, "", fmt.Errorf("%w: nama key wajib diisi", domain.ErrInvalid)
	}
	token, err := apikey.Generate()
	if err != nil {
		return domain.APIKey{}, "", err
	}
	k, err := s.repo.Create(ctx, domain.APIKey{
		UserID:    userID,
		Name:      strings.TrimSpace(name),
		KeyHash:   apikey.Hash(token),
		KeyPrefix: apikey.DisplayPrefix(token),
	})
	if err != nil {
		return domain.APIKey{}, "", err
	}
	return k, token, nil
}

// Authenticate resolves a bearer token to its owning user, recording use.
func (s *APIKeyService) Authenticate(ctx context.Context, token string) (domain.User, error) {
	k, err := s.repo.ByHash(ctx, apikey.Hash(token))
	if err != nil {
		return domain.User{}, domain.ErrUnauthorized
	}
	_ = s.repo.Touch(ctx, k.ID)
	return s.users.ByID(ctx, k.UserID)
}
