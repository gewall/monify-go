package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/alginugraha/monify/internal/domain"
)

type AccountService struct {
	repo AccountRepo
}

func NewAccountService(repo AccountRepo) *AccountService {
	return &AccountService{repo: repo}
}

// AccountInput is the data accepted from a form for create/update.
type AccountInput struct {
	Name           string
	Type           string
	InitialBalance int64
	Archived       bool
}

func (in AccountInput) validate() error {
	if strings.TrimSpace(in.Name) == "" {
		return fmt.Errorf("%w: nama akun wajib diisi", domain.ErrInvalid)
	}
	if !domain.AccountType(in.Type).Valid() {
		return fmt.Errorf("%w: tipe akun tidak valid", domain.ErrInvalid)
	}
	return nil
}

func (s *AccountService) List(ctx context.Context) ([]domain.Account, error) {
	return s.repo.List(ctx)
}

func (s *AccountService) Create(ctx context.Context, in AccountInput) (domain.Account, error) {
	if err := in.validate(); err != nil {
		return domain.Account{}, err
	}
	return s.repo.Create(ctx, domain.Account{
		Name:           strings.TrimSpace(in.Name),
		Type:           domain.AccountType(in.Type),
		InitialBalance: in.InitialBalance,
		Currency:       "IDR",
	})
}

func (s *AccountService) Update(ctx context.Context, id string, in AccountInput) error {
	if err := in.validate(); err != nil {
		return err
	}
	cur, err := s.repo.ByID(ctx, id)
	if err != nil {
		return err
	}
	cur.Name = strings.TrimSpace(in.Name)
	cur.Type = domain.AccountType(in.Type)
	cur.InitialBalance = in.InitialBalance
	cur.Archived = in.Archived
	return s.repo.Update(ctx, cur)
}

func (s *AccountService) Delete(ctx context.Context, id string) error {
	return s.repo.SoftDelete(ctx, id)
}
