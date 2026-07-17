package usecase

import "testing"

func TestLastNonEmptyLine(t *testing.T) {
	cases := map[string]string{
		"linha1\nlinha2\nlinha3":  "linha3",
		"linha1\nlinha2\n\n   \n": "linha2",
		"única linha":             "única linha",
		"":                        "",
		"\n\n\n":                  "",
		"  espaços  \n":           "espaços",
	}
	for in, want := range cases {
		if got := lastNonEmptyLine(in); got != want {
			t.Errorf("lastNonEmptyLine(%q) = %q, want %q", in, got, want)
		}
	}
}
