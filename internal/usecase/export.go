package usecase

import (
	"fmt"
	"time"

	"telegram-career-bot/internal/domain"
	"telegram-career-bot/internal/port"
)

// Export lê o markdown do career-ops (fonte da verdade) e escreve o
// snapshot local (data/*.json) que os comandos de leitura consultam.
type Export struct {
	Source port.CareerOpsSource
	Store  port.SnapshotStore
}

func (e Export) Run() error {
	apps, err := e.Source.ParseApplications()
	if err != nil {
		return fmt.Errorf("tracker: %w", err)
	}

	pipeline, err := e.Source.ParsePipeline()
	if err != nil {
		return fmt.Errorf("pipeline: %w", err)
	}

	followUps, err := e.Source.ParseFollowUps()
	if err != nil {
		return fmt.Errorf("followups: %w", err)
	}

	reports, err := e.Source.ParseReports()
	if err != nil {
		return fmt.Errorf("reports: %w", err)
	}

	stats := ComputeStats(apps)
	now := time.Now().Format("2006-01-02T15:04:05")

	pipeline.ExportedAt = now
	stats.ExportedAt = now

	snapshot := port.Snapshot{
		Tracker:   domain.TrackerData{Applications: apps, ExportedAt: now},
		Pipeline:  *pipeline,
		FollowUps: domain.FollowUpData{Items: followUps, ExportedAt: now},
		Reports:   domain.ReportsIndexData{Reports: reports, ExportedAt: now},
		Stats:     stats,
	}

	return e.Store.WriteAll(snapshot)
}
