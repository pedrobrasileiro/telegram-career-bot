package jsonstore

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"telegram-career-bot/internal/domain"
	"telegram-career-bot/internal/port"
)

// Store implementa port.SnapshotStore lendo/escrevendo data/*.json.
type Store struct {
	DataPath string
}

func New(dataPath string) Store {
	return Store{DataPath: dataPath}
}

func (s Store) LoadTracker() (*domain.TrackerData, error) {
	return loadJSON[domain.TrackerData](filepath.Join(s.DataPath, "tracker.json"))
}

func (s Store) LoadPipeline() (*domain.Pipeline, error) {
	return loadJSON[domain.Pipeline](filepath.Join(s.DataPath, "pipeline.json"))
}

func (s Store) LoadFollowUps() (*domain.FollowUpData, error) {
	return loadJSON[domain.FollowUpData](filepath.Join(s.DataPath, "followups.json"))
}

func (s Store) LoadReports() (*domain.ReportsIndexData, error) {
	return loadJSON[domain.ReportsIndexData](filepath.Join(s.DataPath, "reports-index.json"))
}

func (s Store) LoadStats() (*domain.Stats, error) {
	return loadJSON[domain.Stats](filepath.Join(s.DataPath, "stats.json"))
}

func (s Store) WriteAll(snap port.Snapshot) error {
	if err := writeJSON(filepath.Join(s.DataPath, "tracker.json"), snap.Tracker); err != nil {
		return fmt.Errorf("tracker.json: %w", err)
	}
	if err := writeJSON(filepath.Join(s.DataPath, "pipeline.json"), snap.Pipeline); err != nil {
		return fmt.Errorf("pipeline.json: %w", err)
	}
	if err := writeJSON(filepath.Join(s.DataPath, "followups.json"), snap.FollowUps); err != nil {
		return fmt.Errorf("followups.json: %w", err)
	}
	if err := writeJSON(filepath.Join(s.DataPath, "reports-index.json"), snap.Reports); err != nil {
		return fmt.Errorf("reports-index.json: %w", err)
	}
	if err := writeJSON(filepath.Join(s.DataPath, "stats.json"), snap.Stats); err != nil {
		return fmt.Errorf("stats.json: %w", err)
	}
	return nil
}

func loadJSON[T any](path string) (*T, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var result T
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func writeJSON(path string, v interface{}) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
