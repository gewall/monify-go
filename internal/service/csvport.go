package service

import (
	"encoding/csv"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/alginugraha/monify/internal/domain"
	"github.com/alginugraha/monify/internal/platform/money"
)

// CSVRow is one parsed line from an import file (raw string values).
type CSVRow struct {
	LineNo    int
	Date      string
	Kind      string
	Amount    string
	Account   string
	ToAccount string
	Category  string
	Note      string
}

// csvHeader is the canonical column order for both import and export.
var csvHeader = []string{"date", "kind", "amount", "account", "to_account", "category", "note"}

// ParseTransactionCSV reads a CSV with the canonical header and returns raw rows.
func ParseTransactionCSV(r io.Reader) ([]CSVRow, error) {
	cr := csv.NewReader(r)
	cr.FieldsPerRecord = -1
	cr.TrimLeadingSpace = true

	records, err := cr.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrInvalid, err)
	}
	if len(records) == 0 {
		return nil, fmt.Errorf("%w: file kosong", domain.ErrInvalid)
	}

	idx := map[string]int{}
	for i, h := range records[0] {
		idx[strings.ToLower(strings.TrimSpace(h))] = i
	}
	for _, want := range []string{"date", "kind", "amount", "account"} {
		if _, ok := idx[want]; !ok {
			return nil, fmt.Errorf("%w: kolom wajib '%s' tidak ada", domain.ErrInvalid, want)
		}
	}

	get := func(rec []string, name string) string {
		if i, ok := idx[name]; ok && i < len(rec) {
			return strings.TrimSpace(rec[i])
		}
		return ""
	}

	var out []CSVRow
	for i, rec := range records[1:] {
		if len(rec) == 0 || strings.TrimSpace(strings.Join(rec, "")) == "" {
			continue
		}
		out = append(out, CSVRow{
			LineNo:    i + 2,
			Date:      get(rec, "date"),
			Kind:      get(rec, "kind"),
			Amount:    get(rec, "amount"),
			Account:   get(rec, "account"),
			ToAccount: get(rec, "to_account"),
			Category:  get(rec, "category"),
			Note:      get(rec, "note"),
		})
	}
	return out, nil
}

// DedupKey builds the key used to spot already-imported transactions.
func DedupKey(kind string, amountMinor int64, date time.Time, note string) string {
	return fmt.Sprintf("%s|%d|%s|%s", kind, amountMinor,
		date.Format("2006-01-02"), strings.ToLower(strings.TrimSpace(note)))
}

// WriteTransactionCSV writes transactions in the canonical format.
func WriteTransactionCSV(w io.Writer, txns []domain.Transaction) error {
	cw := csv.NewWriter(w)
	if err := cw.Write(csvHeader); err != nil {
		return err
	}
	for _, t := range txns {
		rec := []string{
			t.OccurredAt.Format("2006-01-02"),
			string(t.Kind),
			money.Format(t.AmountMinor),
			t.AccountName,
			t.ToAccountName,
			t.CategoryName,
			t.Note,
		}
		if err := cw.Write(rec); err != nil {
			return err
		}
	}
	cw.Flush()
	return cw.Error()
}
