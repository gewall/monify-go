package domain

import "time"

// Budget is a spending cap for one category in one month.
type Budget struct {
	ID          string
	CategoryID  string
	Period      time.Time // first day of month
	AmountMinor int64
}

// BudgetLine is a category with its budget and actual spend for a month.
type BudgetLine struct {
	CategoryID   string
	CategoryName string
	AmountMinor  int64 // budgeted (0 = not set)
	SpentMinor   int64
}

// CategorySpend is an aggregated spend total for one category.
type CategorySpend struct {
	CategoryID   string
	CategoryName string
	AmountMinor  int64
}

// DashboardSummary is the data behind the dashboard widgets.
type DashboardSummary struct {
	Month        time.Time
	TotalBalance int64
	MonthIncome  int64
	MonthExpense int64
	TopExpenses  []CategorySpend
	Budgets      []BudgetLine
}

// MonthNet returns income minus expense for the month.
func (s DashboardSummary) MonthNet() int64 { return s.MonthIncome - s.MonthExpense }
