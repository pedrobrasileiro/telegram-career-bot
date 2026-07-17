package opencode

import "testing"

func TestLogWriterAcumulaTudo(t *testing.T) {
	lw := &logWriter{}

	lw.Write([]byte("linha1\nlinha2\n"))
	lw.Write([]byte("sem quebra no fim"))
	lw.flush()

	got := lw.buf.String()
	want := "linha1\nlinha2\nsem quebra no fim"
	if got != want {
		t.Errorf("buf = %q, want %q", got, want)
	}
}

func TestLogWriterFuncionaComEscritasFragmentadas(t *testing.T) {
	lw := &logWriter{}

	// Simula o SO entregando os bytes em pedaços, sem respeitar linhas.
	lw.Write([]byte("li"))
	lw.Write([]byte("nha1\nli"))
	lw.Write([]byte("nha2\n"))
	lw.flush()

	got := lw.buf.String()
	want := "linha1\nlinha2\n"
	if got != want {
		t.Errorf("buf = %q, want %q", got, want)
	}
}

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
