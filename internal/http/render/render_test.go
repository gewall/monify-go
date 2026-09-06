package render

import "testing"

func TestNewParsesAllTemplates(t *testing.T) {
	r, err := New()
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	if _, ok := r.pages["login.html"]; !ok {
		t.Fatal("login.html page not registered")
	}
	if _, ok := r.pages["dashboard.html"]; !ok {
		t.Fatal("dashboard.html page not registered")
	}
}

func TestRupiah(t *testing.T) {
	cases := map[int64]string{
		0:         "Rp0",
		100:       "Rp1",
		123456789: "Rp1.234.567",
		-500000:   "-Rp5.000",
	}
	for in, want := range cases {
		if got := funcs["rupiah"].(func(int64) string)(in); got != want {
			t.Errorf("rupiah(%d) = %q, want %q", in, got, want)
		}
	}
}
