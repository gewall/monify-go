package money

import "testing"

func TestParse(t *testing.T) {
	cases := map[string]int64{
		"150000":     15000000,
		"150.000":    15000000,
		"150.000,50": 15000050,
		"1500,5":     150050,
		"0":          0,
		"-5000":      -500000,
		"Rp 10.000":  1000000,
	}
	for in, want := range cases {
		got, err := Parse(in)
		if err != nil {
			t.Fatalf("Parse(%q) error: %v", in, err)
		}
		if got != want {
			t.Errorf("Parse(%q) = %d, want %d", in, got, want)
		}
	}
	if _, err := Parse("abc"); err == nil {
		t.Error("Parse(abc) should fail")
	}
}

func TestFormat(t *testing.T) {
	cases := map[int64]string{
		0:         "0",
		15000000:  "150.000",
		123456789: "1.234.567",
		-500000:   "-5.000",
	}
	for in, want := range cases {
		if got := Format(in); got != want {
			t.Errorf("Format(%d) = %q, want %q", in, got, want)
		}
	}
}
