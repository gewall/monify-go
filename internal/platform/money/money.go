// Package money handles amounts as int64 minor units (sen). Never use float.
package money

import (
	"errors"
	"strings"
)

// Parse converts a human-typed amount into minor units (sen).
// Accepts "150000", "150.000", "150.000,50", "1500,5". "." is treated as a
// thousands separator, "," as the decimal separator.
func Parse(s string) (int64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, errors.New("empty amount")
	}
	neg := strings.HasPrefix(s, "-")
	s = strings.TrimPrefix(strings.TrimPrefix(s, "-"), "Rp")
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, ".", "")

	whole, frac := s, ""
	if i := strings.IndexByte(s, ','); i >= 0 {
		whole, frac = s[:i], s[i+1:]
	}
	if whole == "" {
		whole = "0"
	}

	var cents int64
	for _, r := range whole {
		if r < '0' || r > '9' {
			return 0, errors.New("invalid amount: " + s)
		}
		cents = cents*10 + int64(r-'0')
	}
	cents *= 100

	switch len(frac) {
	case 0:
		// nothing
	case 1:
		if frac[0] < '0' || frac[0] > '9' {
			return 0, errors.New("invalid fraction")
		}
		cents += int64(frac[0]-'0') * 10
	default:
		for _, r := range frac[:2] {
			if r < '0' || r > '9' {
				return 0, errors.New("invalid fraction")
			}
		}
		cents += int64(frac[0]-'0')*10 + int64(frac[1]-'0')
	}

	if neg {
		cents = -cents
	}
	return cents, nil
}

// Format renders minor units as "1.234.567" (no currency prefix).
func Format(minor int64) string {
	neg := minor < 0
	if neg {
		minor = -minor
	}
	whole := minor / 100
	digits := itoa(whole)
	var b strings.Builder
	if neg {
		b.WriteByte('-')
	}
	for i, d := range []byte(digits) {
		if i > 0 && (len(digits)-i)%3 == 0 {
			b.WriteByte('.')
		}
		b.WriteByte(d)
	}
	return b.String()
}

func itoa(n int64) string {
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}
