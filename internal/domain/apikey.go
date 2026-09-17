package domain

import "time"

// APIKey authenticates an external client acting on behalf of a user.
// Only KeyHash is ever persisted; the plaintext token is shown once at
// creation time.
type APIKey struct {
	ID         string
	UserID     string
	Name       string
	KeyHash    string
	KeyPrefix  string
	LastUsedAt *time.Time
	CreatedAt  time.Time
	RevokedAt  *time.Time
}
