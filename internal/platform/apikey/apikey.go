// Package apikey generates and hashes bearer tokens for external API access.
package apikey

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
)

const tokenPrefix = "mnfy_"

// Generate returns a new random plaintext token. It is shown to the caller
// once; only its hash is ever stored.
func Generate() (string, error) {
	b := make([]byte, 24)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return tokenPrefix + hex.EncodeToString(b), nil
}

// Hash returns the storage/lookup form of a plaintext token.
func Hash(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

// DisplayPrefix returns a short, non-secret slice of the token suitable for
// identifying a key in a listing (e.g. "mnfy_ab12cd34").
func DisplayPrefix(token string) string {
	const n = len(tokenPrefix) + 8
	if len(token) < n {
		return token
	}
	return token[:n]
}
