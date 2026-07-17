package usecase

import "testing"

func TestExtractScanSummaryComResultado(t *testing.T) {
	out := "log inicial\nPortal Scan concluído\nNew added: 3\noutras linhas depois"
	got := extractScanSummary(out)
	if got != "Portal Scan concluído\nNew added: 3\noutras linhas depois" {
		t.Errorf("got %q", got)
	}
}

func TestExtractScanSummarySemMarcador(t *testing.T) {
	got := extractScanSummary("linha sem nenhum marcador conhecido")
	if got != "Scan concluído. Confira o pipeline." {
		t.Errorf("got %q, want fallback padrão", got)
	}
}
