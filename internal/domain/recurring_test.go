package domain

import (
	"testing"
	"time"
)

func d(s string) time.Time {
	t, _ := time.Parse("2006-01-02", s)
	return t
}

func TestRecurringAdvance(t *testing.T) {
	dom := 31
	cases := []struct {
		name string
		rule RecurringRule
		from string
		want string
	}{
		{"daily x3", RecurringRule{Freq: FreqDaily, Interval: 3}, "2026-01-01", "2026-01-04"},
		{"weekly x2", RecurringRule{Freq: FreqWeekly, Interval: 2}, "2026-01-01", "2026-01-15"},
		{"monthly x1", RecurringRule{Freq: FreqMonthly, Interval: 1}, "2026-01-15", "2026-02-15"},
		{"monthly clamps day 31", RecurringRule{Freq: FreqMonthly, Interval: 1, DayOfMonth: &dom}, "2026-01-31", "2026-02-28"},
		{"monthly wraps year", RecurringRule{Freq: FreqMonthly, Interval: 2}, "2026-11-10", "2027-01-10"},
	}
	for _, c := range cases {
		if got := c.rule.Advance(d(c.from)); !got.Equal(d(c.want)) {
			t.Errorf("%s: Advance(%s) = %s, want %s", c.name, c.from, got.Format("2006-01-02"), c.want)
		}
	}
}

func TestSignedFor(t *testing.T) {
	to := "acc-b"
	tr := Transaction{Kind: TxTransfer, AmountMinor: 1000, AccountID: "acc-a", ToAccountID: &to}
	if got := tr.SignedFor("acc-a"); got != -1000 {
		t.Errorf("source side = %d, want -1000", got)
	}
	if got := tr.SignedFor("acc-b"); got != 1000 {
		t.Errorf("dest side = %d, want 1000", got)
	}
}
