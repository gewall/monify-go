package domain

import "time"

// AccountType enumerates the kinds of wallets/accounts.
type AccountType string

const (
	AccountCash    AccountType = "cash"
	AccountBank    AccountType = "bank"
	AccountEwallet AccountType = "ewallet"
	AccountCredit  AccountType = "credit"
)

func (t AccountType) Valid() bool {
	switch t {
	case AccountCash, AccountBank, AccountEwallet, AccountCredit:
		return true
	}
	return false
}

type Account struct {
	ID             string
	Name           string
	Type           AccountType
	InitialBalance int64 // minor units
	Currency       string
	Archived       bool
	CreatedAt      time.Time
	UpdatedAt      time.Time

	// Balance is populated by list queries that aggregate transactions.
	Balance int64
}
