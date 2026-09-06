package domain

import "time"

// Freq is the recurrence frequency.
type Freq string

const (
	FreqDaily   Freq = "daily"
	FreqWeekly  Freq = "weekly"
	FreqMonthly Freq = "monthly"
)

func (f Freq) Valid() bool {
	switch f {
	case FreqDaily, FreqWeekly, FreqMonthly:
		return true
	}
	return false
}

// RecurringRule is a template that generates transactions on a schedule.
type RecurringRule struct {
	ID          string
	Name        string
	Kind        TxKind
	AmountMinor int64
	AccountID   string
	ToAccountID *string
	CategoryID  *string
	Note        string
	Freq        Freq
	Interval    int
	DayOfMonth  *int
	NextRunAt   time.Time
	EndAt       *time.Time
	Active      bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Advance returns the next run date after from, per this rule's frequency.
func (r RecurringRule) Advance(from time.Time) time.Time {
	step := r.Interval
	if step < 1 {
		step = 1
	}
	switch r.Freq {
	case FreqDaily:
		return from.AddDate(0, 0, step)
	case FreqWeekly:
		return from.AddDate(0, 0, 7*step)
	case FreqMonthly:
		y, m := from.Year(), int(from.Month())+step
		y += (m - 1) / 12
		m = (m-1)%12 + 1
		day := from.Day()
		if r.DayOfMonth != nil {
			day = *r.DayOfMonth
		}
		last := daysInMonth(y, m)
		if day > last {
			day = last
		}
		return time.Date(y, time.Month(m), day, 0, 0, 0, 0, from.Location())
	}
	return from
}

func daysInMonth(year, month int) int {
	return time.Date(year, time.Month(month)+1, 0, 0, 0, 0, 0, time.UTC).Day()
}

// ToAccountIDStr / CategoryIDStr mirror the Transaction helpers.
func (r RecurringRule) ToAccountIDStr() string {
	if r.ToAccountID == nil {
		return ""
	}
	return *r.ToAccountID
}

func (r RecurringRule) CategoryIDStr() string {
	if r.CategoryID == nil {
		return ""
	}
	return *r.CategoryID
}
