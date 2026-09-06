package render

import (
	"bytes"
	"testing"
	"time"

	"github.com/alginugraha/monify/internal/domain"
	"github.com/alginugraha/monify/internal/service"
)

// TestPartialsExecute renders the trickier fragments with representative data so
// template bugs (bad field refs, pointer %s, missing funcs) fail the build.
func TestPartialsExecute(t *testing.T) {
	r, err := New()
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now()
	to := "acc-2"

	cases := []struct {
		name string
		data any
	}{
		{"account-row", domain.Account{ID: "a1", Name: "BCA", Type: domain.AccountBank, Balance: 100}},
		{"category-row", domain.Category{ID: "c1", Name: "Makan", Kind: domain.CategoryExpense}},
		{"tx-row", domain.Transaction{ID: "t1", Kind: domain.TxTransfer, AmountMinor: 5000,
			AccountID: "acc-1", ToAccountID: &to, OccurredAt: now, AccountName: "BCA", ToAccountName: "Tunai"}},
		{"tx-panel", map[string]any{"Txns": []domain.Transaction{}, "Page": 1, "Pages": 1, "Total": 0}},
		{"budget-row", map[string]any{"CategoryID": "c1", "CategoryName": "Makan",
			"AmountMinor": int64(50000), "SpentMinor": int64(20000), "Month": "2026-09"}},
		{"rule-row", domain.RecurringRule{ID: "r1", Name: "Netflix", Kind: domain.TxExpense,
			AmountMinor: 199, Freq: domain.FreqMonthly, Interval: 1, NextRunAt: now, Active: true}},
		{"wish-list", map[string]any{
			"Mode": "sequential",
			"View": service.WishlistView{
				AvgSurplus: 100000, Sequential: true,
				Estimates: []domain.WishEstimate{{
					Item:      domain.WishlistItem{ID: "w1", Name: "Laptop", TargetAmountMinor: 1000, Status: domain.WishActive},
					Estimable: true, MonthsNeeded: 3, ReadyDate: now,
				}},
			},
		}},
	}
	for _, c := range cases {
		var buf bytes.Buffer
		if err := r.fragments.ExecuteTemplate(&buf, c.name, c.data); err != nil {
			t.Errorf("%s: %v", c.name, err)
		}
	}
}
