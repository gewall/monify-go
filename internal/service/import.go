package service

import (
	"context"
	"sort"
	"strings"
	"time"

	"github.com/alginugraha/monify/internal/domain"
	"github.com/alginugraha/monify/internal/platform/money"
)

// ImportRow is a CSV row resolved against existing accounts/categories, ready to
// preview and (if selected) import.
type ImportRow struct {
	CSVRow
	Kind        string
	AmountMinor int64
	AccountID   string
	ToAccountID string
	CategoryID  string
	OccurredAt  time.Time
	Duplicate   bool
	Err         string
}

// OK reports whether the row can be imported.
func (r ImportRow) OK() bool { return r.Err == "" }

type ImportService struct {
	accounts   AccountRepo
	categories CategoryRepo
	txns       TransactionRepo
}

func NewImportService(a AccountRepo, c CategoryRepo, t TransactionRepo) *ImportService {
	return &ImportService{accounts: a, categories: c, txns: t}
}

var dateLayouts = []string{"2006-01-02", "02/01/2006", "01/02/2006", "2006/01/02", "02-01-2006"}

func parseCSVDate(s string) (time.Time, bool) {
	for _, l := range dateLayouts {
		if t, err := time.Parse(l, s); err == nil {
			return t, true
		}
	}
	return time.Time{}, false
}

// Preview resolves rows and flags duplicates against existing transactions.
func (s *ImportService) Preview(ctx context.Context, rows []CSVRow) ([]ImportRow, error) {
	accs, err := s.accounts.List(ctx)
	if err != nil {
		return nil, err
	}
	cats, err := s.categories.List(ctx)
	if err != nil {
		return nil, err
	}
	accByName := map[string]string{}
	for _, a := range accs {
		accByName[strings.ToLower(a.Name)] = a.ID
	}
	catByName := map[string]string{}
	for _, c := range cats {
		catByName[strings.ToLower(c.Name)] = c.ID
	}

	out := make([]ImportRow, 0, len(rows))
	var minD, maxD time.Time
	for _, raw := range rows {
		ir := ImportRow{CSVRow: raw}

		d, ok := parseCSVDate(raw.Date)
		if !ok {
			ir.Err = "tanggal tidak dikenali: " + raw.Date
			out = append(out, ir)
			continue
		}
		ir.OccurredAt = d
		if minD.IsZero() || d.Before(minD) {
			minD = d
		}
		if maxD.IsZero() || d.After(maxD) {
			maxD = d
		}

		kind := strings.ToLower(raw.Kind)
		if !domain.TxKind(kind).Valid() {
			ir.Err = "jenis tidak valid: " + raw.Kind
			out = append(out, ir)
			continue
		}
		ir.Kind = kind

		amt, err := money.Parse(raw.Amount)
		if err != nil || amt <= 0 {
			ir.Err = "jumlah tidak valid: " + raw.Amount
			out = append(out, ir)
			continue
		}
		ir.AmountMinor = amt

		id, ok := accByName[strings.ToLower(raw.Account)]
		if !ok {
			ir.Err = "akun tidak ditemukan: " + raw.Account
			out = append(out, ir)
			continue
		}
		ir.AccountID = id

		if kind == string(domain.TxTransfer) {
			toID, ok := accByName[strings.ToLower(raw.ToAccount)]
			if !ok {
				ir.Err = "akun tujuan tidak ditemukan: " + raw.ToAccount
				out = append(out, ir)
				continue
			}
			ir.ToAccountID = toID
		} else if raw.Category != "" {
			cid, ok := catByName[strings.ToLower(raw.Category)]
			if !ok {
				ir.Err = "kategori tidak ditemukan: " + raw.Category
				out = append(out, ir)
				continue
			}
			ir.CategoryID = cid
		}
		out = append(out, ir)
	}

	// Duplicate detection against existing transactions in the date span.
	if !minD.IsZero() {
		to := maxD.AddDate(0, 0, 1)
		existing, err := s.txns.List(ctx, domain.TxFilter{From: &minD, To: &to})
		if err != nil {
			return nil, err
		}
		seen := make(map[string]struct{}, len(existing))
		for _, t := range existing {
			seen[DedupKey(string(t.Kind), t.AmountMinor, t.OccurredAt, t.Note)] = struct{}{}
		}
		for i := range out {
			if !out[i].OK() {
				continue
			}
			k := DedupKey(out[i].Kind, out[i].AmountMinor, out[i].OccurredAt, out[i].Note)
			if _, dup := seen[k]; dup {
				out[i].Duplicate = true
			}
			seen[k] = struct{}{} // also dedupe within the file
		}
	}

	sort.SliceStable(out, func(i, j int) bool { return out[i].LineNo < out[j].LineNo })
	return out, nil
}
