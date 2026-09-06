package domain

import "time"

// TxFilter narrows a transaction listing. Zero values mean "no filter".
type TxFilter struct {
	From       *time.Time
	To         *time.Time
	AccountID  string
	CategoryID string
	Kind       string
	Limit      int
	Offset     int
}
