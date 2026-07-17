package telegram

import "testing"

func TestIsURL(t *testing.T) {
	cases := map[string]bool{
		"https://example.com/vaga":       true,
		"http://example.com":             true,
		"  https://example.com  ":        true,
		"not a url":                      false,
		"":                               false,
		"ftp://example.com":              false,
		"https://example.com com espaço": false,
	}
	for in, want := range cases {
		if got := isURL(in); got != want {
			t.Errorf("isURL(%q) = %v, want %v", in, got, want)
		}
	}
}

func TestParseIntHelper(t *testing.T) {
	n, err := parseInt("  194  ")
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if n != 194 {
		t.Errorf("got %d, want 194", n)
	}

	if _, err := parseInt("abc"); err == nil {
		t.Error("esperava erro para input não numérico")
	}
}

func TestFilepathFromCareerOps(t *testing.T) {
	got := filepathFromCareerOps("/tmp/career-ops", "reports")
	want := "/tmp/career-ops/reports"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
