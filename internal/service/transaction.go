package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/alginugraha/monify/internal/domain"
)

// TransactionRepo is the persistence port for transactions.
type TransactionRepo interface {
	List(ctx context.Context, f domain.TxFilter) ([]domain.Transaction, error)
	Count(ctx context.Context, f domain.TxFilter) (int, error)
	ByID(ctx context.Context, id string) (domain.Transaction, error)
	Create(ctx context.Context, t domain.Transaction) (string, error)
	Update(ctx context.Context, t domain.Transaction) error
	SoftDelete(ctx context.Context, id string) error
}

type TransactionService struct {
	repo TransactionRepo
}

func NewTransactionService(repo TransactionRepo) *TransactionService {
	return &TransactionService{repo: repo}
}

// TxInput is the form payload for creating/updating a transaction.
type TxInput struct {
	Kind        string
	AmountMinor int64
	AccountID   string
	ToAccountID string
	CategoryID  string
	OccurredAt  time.Time
	Note        string
}

func (in TxInput) toDomain() (domain.Transaction, error) {
	kind := domain.TxKind(in.Kind)
	if !kind.Valid() {
		return domain.Transaction{}, fmt.Errorf("%w: jenis transaksi tidak valid", domain.ErrInvalid)
	}
	if in.AmountMinor <= 0 {
		return domain.Transaction{}, fmt.Errorf("%w: jumlah harus lebih dari 0", domain.ErrInvalid)
	}
	if in.AccountID == "" {
		return domain.Transaction{}, fmt.Errorf("%w: akun wajib dipilih", domain.ErrInvalid)
	}
	if in.OccurredAt.IsZero() {
		return domain.Transaction{}, fmt.Errorf("%w: tanggal wajib diisi", domain.ErrInvalid)
	}

	t := domain.Transaction{
		Kind:        kind,
		AmountMinor: in.AmountMinor,
		AccountID:   in.AccountID,
		OccurredAt:  in.OccurredAt,
		Note:        strings.TrimSpace(in.Note),
	}

	switch kind {
	case domain.TxTransfer:
		if in.ToAccountID == "" {
			return domain.Transaction{}, fmt.Errorf("%w: akun tujuan wajib dipilih", domain.ErrInvalid)
		}
		if in.ToAccountID == in.AccountID {
			return domain.Transaction{}, fmt.Errorf("%w: akun asal dan tujuan tidak boleh sama", domain.ErrInvalid)
		}
		to := in.ToAccountID
		t.ToAccountID = &to
	default:
		if in.CategoryID != "" {
			c := in.CategoryID
			t.CategoryID = &c
		}
	}
	return t, nil
}

func (s *TransactionService) List(ctx context.Context, f domain.TxFilter) ([]domain.Transaction, int, error) {
	items, err := s.repo.List(ctx, f)
	if err != nil {
		return nil, 0, err
	}
	total, err := s.repo.Count(ctx, f)
	if err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (s *TransactionService) Get(ctx context.Context, id string) (domain.Transaction, error) {
	return s.repo.ByID(ctx, id)
}

func (s *TransactionService) Create(ctx context.Context, in TxInput) (domain.Transaction, error) {
	t, err := in.toDomain()
	if err != nil {
		return domain.Transaction{}, err
	}
	id, err := s.repo.Create(ctx, t)
	if err != nil {
		return domain.Transaction{}, err
	}
	return s.repo.ByID(ctx, id)
}

func (s *TransactionService) Update(ctx context.Context, id string, in TxInput) (domain.Transaction, error) {
	t, err := in.toDomain()
	if err != nil {
		return domain.Transaction{}, err
	}
	t.ID = id
	if err := s.repo.Update(ctx, t); err != nil {
		return domain.Transaction{}, err
	}
	return s.repo.ByID(ctx, id)
}

func (s *TransactionService) Delete(ctx context.Context, id string) error {
	return s.repo.SoftDelete(ctx, id)
}
