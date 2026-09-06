package domain

import (
	"math"
	"time"
)

// WishStatus is the lifecycle state of a wishlist item.
type WishStatus string

const (
	WishActive   WishStatus = "active"
	WishBought   WishStatus = "bought"
	WishArchived WishStatus = "archived"
)

func (s WishStatus) Valid() bool {
	switch s {
	case WishActive, WishBought, WishArchived:
		return true
	}
	return false
}

// WishlistItem is something the user plans to buy.
type WishlistItem struct {
	ID                string
	Name              string
	TargetAmountMinor int64
	SavedAmountMinor  int64
	Priority          int
	TargetDate        *time.Time
	URL               string
	Status            WishStatus
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// Remaining is how much still needs to be saved (never negative).
func (w WishlistItem) Remaining() int64 {
	r := w.TargetAmountMinor - w.SavedAmountMinor
	if r < 0 {
		return 0
	}
	return r
}

// ProgressPct is 0..100.
func (w WishlistItem) ProgressPct() int {
	if w.TargetAmountMinor <= 0 {
		return 0
	}
	p := int(w.SavedAmountMinor * 100 / w.TargetAmountMinor)
	if p > 100 {
		return 100
	}
	return p
}

// WishEstimate is the projected affordability of one item.
type WishEstimate struct {
	Item          WishlistItem
	Estimable     bool
	AlreadyFunded bool
	MonthsNeeded  int
	ReadyDate     time.Time
	// FundNeededBefore is the cumulative amount that must be saved before this
	// item in sequential mode (0 in parallel mode).
	CumulativeRemaining int64
}

// EstimateWishlist projects when each active item can be bought given an average
// monthly surplus. In sequential mode items are funded one after another in the
// given order; in parallel mode every item draws on the full surplus at once.
func EstimateWishlist(items []WishlistItem, avgMonthlySurplus int64, sequential bool) []WishEstimate {
	out := make([]WishEstimate, 0, len(items))
	var cumulative int64
	now := time.Now()

	for _, it := range items {
		est := WishEstimate{Item: it}
		rem := it.Remaining()
		if rem == 0 {
			est.AlreadyFunded = true
			est.Estimable = true
			est.ReadyDate = now
			out = append(out, est)
			continue
		}

		basis := rem
		if sequential {
			cumulative += rem
			basis = cumulative
		}
		est.CumulativeRemaining = basis

		if avgMonthlySurplus <= 0 {
			est.Estimable = false
			out = append(out, est)
			continue
		}

		months := int(math.Ceil(float64(basis) / float64(avgMonthlySurplus)))
		if months < 1 {
			months = 1
		}
		est.Estimable = true
		est.MonthsNeeded = months
		est.ReadyDate = now.AddDate(0, months, 0)
		out = append(out, est)
	}
	return out
}
