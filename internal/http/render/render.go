// Package render caches HTML templates and renders full pages or htmx fragments.
package render

import (
	"bytes"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"time"

	"github.com/alginugraha/monify/web"
)

// Renderer holds the parsed template set.
type Renderer struct {
	pages     map[string]*template.Template // page name -> layout+page
	fragments *template.Template            // all partials, addressed by define name
}

var funcs = template.FuncMap{
	"rupiah": func(minor int64) string {
		neg := minor < 0
		if neg {
			minor = -minor
		}
		whole := minor / 100
		s := fmt.Sprintf("%d", whole)
		// group thousands
		out := make([]byte, 0, len(s)+len(s)/3)
		for i, d := range []byte(s) {
			if i > 0 && (len(s)-i)%3 == 0 {
				out = append(out, '.')
			}
			out = append(out, d)
		}
		sign := ""
		if neg {
			sign = "-"
		}
		return fmt.Sprintf("%sRp%s", sign, out)
	},
	"date":  func(t time.Time) string { return t.Format("02 Jan 2006") },
	"add":   func(a, b int) int { return a + b },
	"sub64": func(a, b int64) int64 { return a - b },
	"dict": func(kv ...any) map[string]any {
		m := make(map[string]any, len(kv)/2)
		for i := 0; i+1 < len(kv); i += 2 {
			key, _ := kv[i].(string)
			m[key] = kv[i+1]
		}
		return m
	},
	"pct": func(part, whole int64) int {
		if whole <= 0 {
			return 0
		}
		p := int(part * 100 / whole)
		if p < 0 {
			return 0
		}
		return p
	},
}

// New parses every template from the embedded filesystem.
func New() (*Renderer, error) {
	r := &Renderer{pages: make(map[string]*template.Template)}

	pageFiles, err := fs.Glob(web.FS, "template/page/*.html")
	if err != nil {
		return nil, err
	}
	for _, pf := range pageFiles {
		name := baseName(pf)
		t, err := template.New("base.html").Funcs(funcs).ParseFS(web.FS,
			"template/layout/base.html", "template/partial/*.html", pf)
		if err != nil {
			return nil, fmt.Errorf("parse page %s: %w", name, err)
		}
		r.pages[name] = t
	}

	frag, err := template.New("fragments").Funcs(funcs).ParseFS(web.FS, "template/partial/*.html")
	if err != nil {
		return nil, err
	}
	r.fragments = frag
	return r, nil
}

// Page renders a full HTML page through the base layout.
func (r *Renderer) Page(w http.ResponseWriter, status int, name string, data any) {
	t, ok := r.pages[name]
	if !ok {
		http.Error(w, "unknown page: "+name, http.StatusInternalServerError)
		return
	}
	r.exec(w, status, t, "base.html", data)
}

// Partial renders a single {{define}} block, for htmx swaps.
func (r *Renderer) Partial(w http.ResponseWriter, status int, define string, data any) {
	r.exec(w, status, r.fragments, define, data)
}

func (r *Renderer) exec(w http.ResponseWriter, status int, t *template.Template, name string, data any) {
	var buf bytes.Buffer
	if err := t.ExecuteTemplate(&buf, name, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	_, _ = buf.WriteTo(w)
}

func baseName(path string) string {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' {
			return path[i+1:]
		}
	}
	return path
}
