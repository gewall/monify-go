package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/alginugraha/monify/internal/domain"
)

// WishlistRepo is the persistence port for wishlist items.
type WishlistRepo interface {
	List(ctx context.Context) ([]domain.WishlistItem, error)
	ByID(ctx context.Context, id string) (domain.WishlistItem, error)
	Create(ctx context.Context, w domain.WishlistItem) (domain.WishlistItem, error)
	Update(ctx context.Context, w domain.WishlistItem) error
	SoftDelete(ctx context.Context, id string) error
}

// SurplusReader supplies the average monthly surplus used for estimates.
type SurplusReader interface {
	AvgMonthlySurplus(ctx context.Context, months int) (int64, error)
}

type WishlistService struct {
	repo    WishlistRepo
	surplus SurplusReader
}

func NewWishlistService(repo WishlistRepo, surplus SurplusReader) *WishlistService {
	return &WishlistService{repo: repo, surplus: surplus}
}

type WishInput struct {
	Name        string
	TargetMinor int64
	SavedMinor  int64
	Priority    int
	TargetDate  *time.Time
	URL         string
	Status      string
}

func (in WishInput) toDomain() (domain.WishlistItem, error) {
	if strings.TrimSpace(in.Name) == "" {
		return domain.WishlistItem{}, fmt.Errorf("%w: nama wajib diisi", domain.ErrInvalid)
	}
	if in.TargetMinor <= 0 {
		return domain.WishlistItem{}, fmt.Errorf("%w: harga target harus > 0", domain.ErrInvalid)
	}
	if in.SavedMinor < 0 {
		return domain.WishlistItem{}, fmt.Errorf("%w: tabungan tidak boleh negatif", domain.ErrInvalid)
	}
	status := domain.WishStatus(in.Status)
	if in.Status == "" {
		status = domain.WishActive
	}
	if !status.Valid() {
		return domain.WishlistItem{}, fmt.Errorf("%w: status tidak valid", domain.ErrInvalid)
	}
	prio := in.Priority
	if prio <= 0 {
		prio = 100
	}
	return domain.WishlistItem{
		Name:              strings.TrimSpace(in.Name),
		TargetAmountMinor: in.TargetMinor,
		SavedAmountMinor:  in.SavedMinor,
		Priority:          prio,
		TargetDate:        in.TargetDate,
		URL:               strings.TrimSpace(in.URL),
		Status:            status,
	}, nil
}

// WishlistView is the wishlist plus its estimates for a chosen mode.
type WishlistView struct {
	Estimates  []domain.WishEstimate
	AvgSurplus int64
	Sequential bool
}

func (s *WishlistService) Overview(ctx context.Context, sequential bool) (WishlistView, error) {
	items, err := s.repo.List(ctx)
	if err != nil {
		return WishlistView{}, err
	}
	avg, err := s.surplus.AvgMonthlySurplus(ctx, 3)
	if err != nil {
		return WishlistView{}, err
	}

	// Only active items feed the sequential cumulative; bought/archived shown as-is.
	active := make([]domain.WishlistItem, 0, len(items))
	rest := make([]domain.WishlistItem, 0)
	for _, it := range items {
		if it.Status == domain.WishActive {
			active = append(active, it)
		} else {
			rest = append(rest, it)
		}
	}

	est := domain.EstimateWishlist(active, avg, sequential)
	for _, it := range rest {
		est = append(est, domain.WishEstimate{Item: it, Estimable: false})
	}
	return WishlistView{Estimates: est, AvgSurplus: avg, Sequential: sequential}, nil
}

func (s *WishlistService) Get(ctx context.Context, id string) (domain.WishlistItem, error) {
	return s.repo.ByID(ctx, id)
}

func (s *WishlistService) Create(ctx context.Context, in WishInput) (domain.WishlistItem, error) {
	w, err := in.toDomain()
	if err != nil {
		return domain.WishlistItem{}, err
	}
	return s.repo.Create(ctx, w)
}

func (s *WishlistService) Update(ctx context.Context, id string, in WishInput) (domain.WishlistItem, error) {
	w, err := in.toDomain()
	if err != nil {
		return domain.WishlistItem{}, err
	}
	w.ID = id
	if err := s.repo.Update(ctx, w); err != nil {
		return domain.WishlistItem{}, err
	}
	return s.repo.ByID(ctx, id)
}

// AddSaving increments the saved amount (delta may be negative to correct).
func (s *WishlistService) AddSaving(ctx context.Context, id string, delta int64) (domain.WishlistItem, error) {
	w, err := s.repo.ByID(ctx, id)
	if err != nil {
		return domain.WishlistItem{}, err
	}
	w.SavedAmountMinor += delta
	if w.SavedAmountMinor < 0 {
		w.SavedAmountMinor = 0
	}
	if err := s.repo.Update(ctx, w); err != nil {
		return domain.WishlistItem{}, err
	}
	return w, nil
}

func (s *WishlistService) SetStatus(ctx context.Context, id string, status domain.WishStatus) (domain.WishlistItem, error) {
	if !status.Valid() {
		return domain.WishlistItem{}, fmt.Errorf("%w: status tidak valid", domain.ErrInvalid)
	}
	w, err := s.repo.ByID(ctx, id)
	if err != nil {
		return domain.WishlistItem{}, err
	}
	w.Status = status
	if err := s.repo.Update(ctx, w); err != nil {
		return domain.WishlistItem{}, err
	}
	return w, nil
}

func (s *WishlistService) Delete(ctx context.Context, id string) error {
	return s.repo.SoftDelete(ctx, id)
}
