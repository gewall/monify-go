// Package apihandler implements the token-authenticated JSON REST API used
// by external applications to integrate with Monify (see docs/API.md).
// Handlers here are a thin JSON-encoding layer over the same internal/service
// layer the htmx handlers use — no business logic lives in this package.
package apihandler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/alginugraha/monify/internal/domain"
)

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	msg := "internal error"
	switch {
	case errors.Is(err, domain.ErrNotFound):
		status, msg = http.StatusNotFound, "not found"
	case errors.Is(err, domain.ErrInvalid):
		status, msg = http.StatusBadRequest, err.Error()
	case errors.Is(err, domain.ErrConflict):
		status, msg = http.StatusConflict, err.Error()
	case errors.Is(err, domain.ErrUnauthorized):
		status, msg = http.StatusUnauthorized, "unauthorized"
	}
	writeJSON(w, status, map[string]string{"error": msg})
}

func errJoinInvalid(err error) error {
	return errors.Join(domain.ErrInvalid, err)
}

func decodeJSON(r *http.Request, v any) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		return errors.Join(domain.ErrInvalid, err)
	}
	return nil
}
