package usecase

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"telegram-career-bot/internal/domain"
	"telegram-career-bot/internal/port"
)

// Ask responde uma pergunta em linguagem natural sobre os dados do
// tracker, usando o opencode como LLM com o contexto injetado no prompt.
type Ask struct {
	Runner        port.JobRunner
	Store         port.SnapshotStore
	CareerOpsPath string
	Timeout       time.Duration
}

func (a Ask) Run(ctx context.Context, question string) (string, error) {
	dataCtx := a.buildDataContext()
	if dataCtx == "" {
		dataCtx = "(dados ainda não exportados — rode go run ./cmd/bot sync)"
	}

	prompt := fmt.Sprintf("Você é um assistente de job search. Responda a pergunta do usuário em português brasileiro (pt-BR) de forma concisa e direta, usando apenas os dados fornecidos abaixo. Se os dados não forem suficientes para responder, diga exatamente o que falta.\n\n## DADOS DO TRACKER\n\n%s\n\n## PERGUNTA DO USUÁRIO\n\n%s", dataCtx, question)

	out, err := a.Runner.Run(ctx, a.CareerOpsPath, prompt, a.Timeout)
	if err != nil {
		return "", err
	}

	answer := strings.TrimSpace(out)
	if len(answer) > 4000 {
		answer = answer[:4000] + "\n\n…"
	}
	return answer, nil
}

func (a Ask) buildDataContext() string {
	tracker, err := a.Store.LoadTracker()
	if err != nil {
		return ""
	}
	stats, err := a.Store.LoadStats()
	if err != nil {
		return ""
	}
	pipeline, _ := a.Store.LoadPipeline()

	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Total de vagas: %d\n", stats.Total))
	sb.WriteString(fmt.Sprintf("Score médio: %.1f/5\n\n", stats.AverageScore))

	sb.WriteString("Por status:\n")
	for _, s := range StatusOrder() {
		if c, ok := stats.ByStatus[s]; ok && c > 0 {
			sb.WriteString(fmt.Sprintf("  %s: %d\n", s, c))
		}
	}

	if pipeline != nil {
		sb.WriteString(fmt.Sprintf("\nPipeline: %d pendentes\n", len(pipeline.Pending)))
	}

	sb.WriteString("\nÚltimas 20 vagas (ordenadas por #):\n")
	sorted := make([]domain.Application, len(tracker.Applications))
	copy(sorted, tracker.Applications)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Num > sorted[j].Num
	})
	limit := 20
	if len(sorted) < limit {
		limit = len(sorted)
	}
	for i := 0; i < limit; i++ {
		a := sorted[i]
		sb.WriteString(fmt.Sprintf("#%d %s | %s | %s | %s", a.Num, a.Date, a.Company, a.Role, a.Status))
		if a.Score != "" && a.Score != "N/A" {
			sb.WriteString(fmt.Sprintf(" | %s", a.Score))
		}
		if a.Notes != "" {
			notes := a.Notes
			if len(notes) > 120 {
				notes = notes[:120] + "..."
			}
			sb.WriteString(fmt.Sprintf(" | %s", notes))
		}
		sb.WriteByte('\n')
	}

	return sb.String()
}
