package domain

import "time"

// CategoryKind is income or expense.
type CategoryKind string

const (
	CategoryIncome  CategoryKind = "income"
	CategoryExpense CategoryKind = "expense"
)

func (k CategoryKind) Valid() bool {
	return k == CategoryIncome || k == CategoryExpense
}

type Category struct {
	ID        string
	Name      string
	Kind      CategoryKind
	ParentID  *string
	CreatedAt time.Time
	UpdatedAt time.Time
}
