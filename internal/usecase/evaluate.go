package usecase

import (
	"context"
	"fmt"
	"strings"
	"time"

	"telegram-career-bot/internal/domain"
	"telegram-career-bot/internal/port"
)

type EvaluateResult struct {
	Report    *domain.ReportSummary // report mais recente, se o índice não estiver vazio
	RawOutput string                // fallback: última linha não-vazia da saída do opencode
}

// Evaluate roda a avaliação de uma vaga via career-ops auto-pipeline,
// reexporta o snapshot e busca o report gerado.
type Evaluate struct {
	Runner        port.JobRunner
	Exporter      Export
	Store         port.SnapshotStore
	CareerOpsPath string
	Timeout       time.Duration
}

func (e Evaluate) Run(ctx context.Context, url string) (*EvaluateResult, error) {
	prompt := fmt.Sprintf("Evaluate this JD with career-ops auto-pipeline: %s", url)
	out, err := e.Runner.Run(ctx, e.CareerOpsPath, prompt, e.Timeout)
	if err != nil {
		return nil, err
	}

	// Reexporta pra refletir o novo report; erro de export não bloqueia
	// a resposta (mesmo comportamento de antes da refatoração).
	_ = e.Exporter.Run()

	reportsData, err := e.Store.LoadReports()
	if err == nil && reportsData != nil && len(reportsData.Reports) > 0 {
		latest := reportsData.Reports[0]
		return &EvaluateResult{Report: &latest}, nil
	}

	return &EvaluateResult{RawOutput: lastNonEmptyLine(out)}, nil
}

func lastNonEmptyLine(s string) string {
	lines := strings.Split(strings.TrimSpace(s), "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		if line := strings.TrimSpace(lines[i]); line != "" {
			return line
		}
	}
	return ""
}
