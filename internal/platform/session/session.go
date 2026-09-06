// Package session implements a stateless, HMAC-signed cookie session.
// Suitable for a single-user app: no server-side store, survives restarts.
package session

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const cookieName = "monify_session"

// Manager signs and verifies session cookies.
type Manager struct {
	key    []byte
	secure bool
	ttl    time.Duration
}

func NewManager(secret string, secure bool) *Manager {
	return &Manager{key: []byte(secret), secure: secure, ttl: 30 * 24 * time.Hour}
}

// Set writes a signed session cookie identifying the user.
func (m *Manager) Set(w http.ResponseWriter, userID string) {
	exp := time.Now().Add(m.ttl).Unix()
	payload := userID + "|" + strconv.FormatInt(exp, 10)
	value := payload + "|" + m.sign(payload)

	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    base64.RawURLEncoding.EncodeToString([]byte(value)),
		Path:     "/",
		HttpOnly: true,
		Secure:   m.secure,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Unix(exp, 0),
	})
}

// Clear removes the session cookie.
func (m *Manager) Clear(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   m.secure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
}

// UserID returns the authenticated user id from the request, if any.
func (m *Manager) UserID(r *http.Request) (string, bool) {
	c, err := r.Cookie(cookieName)
	if err != nil {
		return "", false
	}
	raw, err := base64.RawURLEncoding.DecodeString(c.Value)
	if err != nil {
		return "", false
	}
	userID, exp, err := m.parse(string(raw))
	if err != nil {
		return "", false
	}
	if time.Now().Unix() > exp {
		return "", false
	}
	return userID, true
}

func (m *Manager) parse(value string) (userID string, exp int64, err error) {
	parts := strings.Split(value, "|")
	if len(parts) != 3 {
		return "", 0, errors.New("malformed session")
	}
	payload := parts[0] + "|" + parts[1]
	if !hmac.Equal([]byte(parts[2]), []byte(m.sign(payload))) {
		return "", 0, errors.New("bad signature")
	}
	exp, err = strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return "", 0, fmt.Errorf("bad expiry: %w", err)
	}
	return parts[0], exp, nil
}

func (m *Manager) sign(payload string) string {
	mac := hmac.New(sha256.New, m.key)
	mac.Write([]byte(payload))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}
