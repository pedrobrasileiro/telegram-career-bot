package port

import "telegram-career-bot/internal/domain"

// CareerOpsSource lê os arquivos markdown do repositório career-ops
// (fonte da verdade) e parseia pras entidades de domínio.
type CareerOpsSource interface {
	ParseApplications() ([]domain.Application, error)
	ParsePipeline() (*domain.Pipeline, error)
	ParseFollowUps() ([]domain.FollowUpItem, error)
	ParseReports() ([]domain.ReportSummary, error)
}
