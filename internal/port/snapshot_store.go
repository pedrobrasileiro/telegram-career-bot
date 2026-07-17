package port

import "telegram-career-bot/internal/domain"

// Snapshot é o pacote completo de dados exportado de uma vez (data/*.json).
type Snapshot struct {
	Tracker   domain.TrackerData
	Pipeline  domain.Pipeline
	FollowUps domain.FollowUpData
	Reports   domain.ReportsIndexData
	Stats     domain.Stats
}

// SnapshotStore lê/escreve o snapshot local (data/*.json) usado pelos
// comandos de leitura do bot. Load* retorna erro se o snapshot ainda não
// foi exportado (ex: "go run ./cmd/bot sync" nunca rodou).
type SnapshotStore interface {
	LoadTracker() (*domain.TrackerData, error)
	LoadPipeline() (*domain.Pipeline, error)
	LoadFollowUps() (*domain.FollowUpData, error)
	LoadReports() (*domain.ReportsIndexData, error)
	LoadStats() (*domain.Stats, error)
	WriteAll(Snapshot) error
}
