package handler

import (
	"errors"
	"net/http"

	"github.com/alginugraha/monify/internal/domain"
)

// httpError maps a domain error to an HTTP response. For htmx requests the
// message is returned as a small fragment so it can be swapped into a target.
func httpError(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	msg := "Terjadi kesalahan."
	switch {
	case errors.Is(err, domain.ErrNotFound):
		status, msg = http.StatusNotFound, "Data tidak ditemukan."
	case errors.Is(err, domain.ErrInvalid):
		status, msg = http.StatusBadRequest, err.Error()
	case errors.Is(err, domain.ErrConflict):
		status, msg = http.StatusConflict, err.Error()
	case errors.Is(err, domain.ErrUnauthorized):
		status, msg = http.StatusUnauthorized, "Tidak diizinkan."
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	_, _ = w.Write([]byte(`<p class="error">` + htmlEscape(msg) + `</p>`))
}

func errJoinInvalid(err error) error {
	return errors.Join(domain.ErrInvalid, err)
}

func htmlEscape(s string) string {
	repl := map[rune]string{'<': "&lt;", '>': "&gt;", '&': "&amp;", '"': "&#34;"}
	out := make([]rune, 0, len(s))
	for _, r := range s {
		if e, ok := repl[r]; ok {
			out = append(out, []rune(e)...)
			continue
		}
		out = append(out, r)
	}
	return string(out)
}
