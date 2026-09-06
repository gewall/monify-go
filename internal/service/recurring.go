package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/alginugraha/monify/internal/domain"
)

// RecurringRepo is the persistence port for recurring rules.
type RecurringRepo interface {
	List(ctx context.Context) ([]domain.RecurringRule, error)
	ByID(ctx context.Context, id string) (domain.RecurringRule, error)
	Create(ctx context.Context, r domain.RecurringRule) (domain.RecurringRule, error)
	Update(ctx context.Context, r domain.RecurringRule) error
	SoftDelete(ctx context.Context, id string) error
	RunDue(ctx context.Context, asOf time.Time) (int, error)
}

type RecurringService struct {
	repo RecurringRepo
}

func NewRecurringService(repo RecurringRepo) *RecurringService {
	return &RecurringService{repo: repo}
}

// RuleInput is the form payload for a recurring rule.
type RuleInput struct {
	Name        string
	Kind        string
	AmountMinor int64
	AccountID   string
	ToAccountID string
	CategoryID  string
	Note        string
	Freq        string
	Interval    int
	DayOfMonth  int
	NextRunAt   time.Time
	EndAt       *time.Time
	Active      bool
}

func (in RuleInput) toDomain() (domain.RecurringRule, error) {
	if strings.TrimSpace(in.Name) == "" {
		return domain.RecurringRule{}, fmt.Errorf("%w: nama wajib diisi", domain.ErrInvalid)
	}
	kind := domain.TxKind(in.Kind)
	if !kind.Valid() {
		return domain.RecurringRule{}, fmt.Errorf("%w: jenis tidak valid", domain.ErrInvalid)
	}
	freq := domain.Freq(in.Freq)
	if !freq.Valid() {
		return domain.RecurringRule{}, fmt.Errorf("%w: frekuensi tidak valid", domain.ErrInvalid)
	}
	if in.AmountMinor <= 0 {
		return domain.RecurringRule{}, fmt.Errorf("%w: jumlah harus > 0", domain.ErrInvalid)
	}
	if in.AccountID == "" {
		return domain.RecurringRule{}, fmt.Errorf("%w: akun wajib dipilih", domain.ErrInvalid)
	}
	if in.NextRunAt.IsZero() {
		return domain.RecurringRule{}, fmt.Errorf("%w: tanggal mulai wajib diisi", domain.ErrInvalid)
	}
	interval := in.Interval
	if interval < 1 {
		interval = 1
	}

	rule := domain.RecurringRule{
		Name:        strings.TrimSpace(in.Name),
		Kind:        kind,
		AmountMinor: in.AmountMinor,
		AccountID:   in.AccountID,
		Note:        strings.TrimSpace(in.Note),
		Freq:        freq,
		Interval:    interval,
		NextRunAt:   in.NextRunAt,
		EndAt:       in.EndAt,
		Active:      in.Active,
	}

	switch kind {
	case domain.TxTransfer:
		if in.ToAccountID == "" || in.ToAccountID == in.AccountID {
			return domain.RecurringRule{}, fmt.Errorf("%w: akun tujuan transfer tidak valid", domain.ErrInvalid)
		}
		to := in.ToAccountID
		rule.ToAccountID = &to
	default:
		if in.CategoryID != "" {
			c := in.CategoryID
			rule.CategoryID = &c
		}
	}
	if freq == domain.FreqMonthly && in.DayOfMonth >= 1 && in.DayOfMonth <= 31 {
		d := in.DayOfMonth
		rule.DayOfMonth = &d
	}
	return rule, nil
}

func (s *RecurringService) List(ctx context.Context) ([]domain.RecurringRule, error) {
	return s.repo.List(ctx)
}

func (s *RecurringService) Get(ctx context.Context, id string) (domain.RecurringRule, error) {
	return s.repo.ByID(ctx, id)
}

func (s *RecurringService) Create(ctx context.Context, in RuleInput) (domain.RecurringRule, error) {
	rule, err := in.toDomain()
	if err != nil {
		return domain.RecurringRule{}, err
	}
	return s.repo.Create(ctx, rule)
}

func (s *RecurringService) Update(ctx context.Context, id string, in RuleInput) (domain.RecurringRule, error) {
	rule, err := in.toDomain()
	if err != nil {
		return domain.RecurringRule{}, err
	}
	rule.ID = id
	if err := s.repo.Update(ctx, rule); err != nil {
		return domain.RecurringRule{}, err
	}
	return s.repo.ByID(ctx, id)
}

func (s *RecurringService) Delete(ctx context.Context, id string) error {
	return s.repo.SoftDelete(ctx, id)
}

// Run generates any due transactions up to now.
func (s *RecurringService) Run(ctx context.Context) (int, error) {
	return s.repo.RunDue(ctx, time.Now())
}

// Upcoming returns active rules due within the next n days, soonest first.
func (s *RecurringService) Upcoming(ctx context.Context, days int) ([]domain.RecurringRule, error) {
	all, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}
	cutoff := time.Now().AddDate(0, 0, days)
	var out []domain.RecurringRule
	for _, r := range all {
		if r.Active && !r.NextRunAt.After(cutoff) {
			out = append(out, r)
		}
	}
	return out, nil
}
