package domain

import "time"

// TxKind is income, expense or transfer.
type TxKind string

const (
	TxIncome   TxKind = "income"
	TxExpense  TxKind = "expense"
	TxTransfer TxKind = "transfer"
)

func (k TxKind) Valid() bool {
	switch k {
	case TxIncome, TxExpense, TxTransfer:
		return true
	}
	return false
}

type Transaction struct {
	ID          string
	Kind        TxKind
	AmountMinor int64
	AccountID   string
	ToAccountID *string
	CategoryID  *string
	OccurredAt  time.Time
	Note        string
	CreatedAt   time.Time
	UpdatedAt   time.Time

	// Display-only fields populated by list joins.
	AccountName   string
	ToAccountName string
	CategoryName  string
}

// ToAccountIDStr returns the destination account id as a plain string ("" if none).
func (t Transaction) ToAccountIDStr() string {
	if t.ToAccountID == nil {
		return ""
	}
	return *t.ToAccountID
}

// CategoryIDStr returns the category id as a plain string ("" if none).
func (t Transaction) CategoryIDStr() string {
	if t.CategoryID == nil {
		return ""
	}
	return *t.CategoryID
}

// SignedFor returns the effect of this transaction on the given account,
// in minor units (positive = money in).
func (t Transaction) SignedFor(accountID string) int64 {
	switch t.Kind {
	case TxIncome:
		return t.AmountMinor
	case TxExpense:
		return -t.AmountMinor
	case TxTransfer:
		if t.ToAccountID != nil && *t.ToAccountID == accountID {
			return t.AmountMinor
		}
		return -t.AmountMinor
	}
	return 0
}
