package main

import (
	"fmt"
	"path/filepath"
	"time"
)

func exportAll(careerOpsPath, dataPath string) error {
	apps, err := parseApplications(careerOpsPath)
	if err != nil {
		return fmt.Errorf("tracker: %w", err)
	}

	pipeline, err := parsePipeline(careerOpsPath)
	if err != nil {
		return fmt.Errorf("pipeline: %w", err)
	}

	followUps, err := parseFollowUps(careerOpsPath)
	if err != nil {
		return fmt.Errorf("followups: %w", err)
	}

	reports, err := parseReportHeaders(careerOpsPath)
	if err != nil {
		return fmt.Errorf("reports: %w", err)
	}

	stats := computeStats(apps)
	now := time.Now().Format("2006-01-02T15:04:05")

	trackerData := TrackerData{
		Applications: apps,
		ExportedAt:   now,
	}
	if err := writeJSON(filepath.Join(dataPath, "tracker.json"), trackerData); err != nil {
		return fmt.Errorf("tracker.json: %w", err)
	}

	if err := writeJSON(filepath.Join(dataPath, "pipeline.json"), struct {
		Pending    []PipelineItem `json:"pending"`
		Processed  []PipelineItem `json:"processed"`
		ExportedAt string         `json:"exportedAt"`
	}{
		Pending:    pipeline.Pending,
		Processed:  pipeline.Processed,
		ExportedAt: now,
	}); err != nil {
		return fmt.Errorf("pipeline.json: %w", err)
	}

	followUpData := FollowUpData{
		Items:      followUps,
		ExportedAt: now,
	}
	if err := writeJSON(filepath.Join(dataPath, "followups.json"), followUpData); err != nil {
		return fmt.Errorf("followups.json: %w", err)
	}

	reportsData := ReportsIndexData{
		Reports:    reports,
		ExportedAt: now,
	}
	if err := writeJSON(filepath.Join(dataPath, "reports-index.json"), reportsData); err != nil {
		return fmt.Errorf("reports-index.json: %w", err)
	}

	stats.ExportedAt = now
	if err := writeJSON(filepath.Join(dataPath, "stats.json"), stats); err != nil {
		return fmt.Errorf("stats.json: %w", err)
	}

	return nil
}

func runSync() {
	careerOpsPath := "../career-ops"
	dataPath := "./data"

	cfg, err := loadConfig()
	if err == nil {
		careerOpsPath = cfg.CareerOpsPath
		dataPath = cfg.DataPath
	}

	fmt.Printf("Exportando dados de %s para %s...\n", careerOpsPath, dataPath)

	if err := exportAll(careerOpsPath, dataPath); err != nil {
		fmt.Printf("Erro: %v\n", err)
		return
	}

	fmt.Println("Export concluído:")
	fmt.Printf("  tracker.json\n  pipeline.json\n  followups.json\n  reports-index.json\n  stats.json\n")
}
