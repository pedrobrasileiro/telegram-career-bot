package usecase

import (
	"telegram-career-bot/internal/domain"
	"telegram-career-bot/internal/port"
)

type PipelineQuery struct {
	Store port.SnapshotStore
}

func (q PipelineQuery) Run() (*domain.Pipeline, error) {
	return q.Store.LoadPipeline()
}
