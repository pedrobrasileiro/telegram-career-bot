package usecase

import (
	"context"
	"strings"
	"time"

	"telegram-career-bot/internal/domain"
	"telegram-career-bot/internal/port"
)

type ScanResult struct {
	Summary  string
	Pipeline *domain.Pipeline
}

// Scan roda o scan completo de vagas (scan.mjs + WebSearch L3) e reexporta
// o snapshot.
type Scan struct {
	Runner        port.JobRunner
	Exporter      Export
	Store         port.SnapshotStore
	CareerOpsPath string
	Timeout       time.Duration
}

func (s Scan) Run(ctx context.Context) (*ScanResult, error) {
	out, err := s.Runner.Run(ctx, s.CareerOpsPath, "Run career-ops scan mode", s.Timeout)
	if err != nil {
		return nil, err
	}

	_ = s.Exporter.Run()

	pipeline, _ := s.Store.LoadPipeline()

	return &ScanResult{Summary: extractScanSummary(out), Pipeline: pipeline}, nil
}

func extractScanSummary(out string) string {
	lines := strings.Split(out, "\n")
	var summaryLines []string
	inSummary := false
	for _, line := range lines {
		if strings.Contains(line, "Portal Scan") || strings.Contains(line, "Offer") || strings.Contains(line, "New added") {
			inSummary = true
		}
		if inSummary {
			summaryLines = append(summaryLines, line)
		}
	}
	if len(summaryLines) == 0 {
		return "Scan concluído. Confira o pipeline."
	}
	return strings.Join(summaryLines, "\n")
}
