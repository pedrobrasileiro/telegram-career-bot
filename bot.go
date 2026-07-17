package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

const logFilePath = "/tmp/telegram-bot.log"

func setupLogging() error {
	f, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("abrindo %s: %w", logFilePath, err)
	}
	log.SetOutput(io.MultiWriter(os.Stdout, f))
	log.SetFlags(log.Ldate | log.Ltime)
	return nil
}

func isURL(text string) bool {
	re := regexp.MustCompile(`^https?://[^\s]+$`)
	return re.MatchString(strings.TrimSpace(text))
}

func parseInt(s string) (int, error) {
	return strconv.Atoi(strings.TrimSpace(s))
}

func filepathFromCareerOps(careerOpsPath, subdir string) string {
	abs, err := filepath.Abs(careerOpsPath)
	if err != nil {
		return careerOpsPath
	}
	return filepath.Join(abs, subdir)
}

func loadTrackerData(dataPath string) *TrackerData {
	return loadJSONFile[TrackerData](filepath.Join(dataPath, "tracker.json"))
}

func loadPipelineData(dataPath string) *struct {
	Pending    []PipelineItem `json:"pending"`
	Processed  []PipelineItem `json:"processed"`
	ExportedAt string         `json:"exportedAt"`
} {
	return loadJSONFile[struct {
		Pending    []PipelineItem `json:"pending"`
		Processed  []PipelineItem `json:"processed"`
		ExportedAt string         `json:"exportedAt"`
	}](filepath.Join(dataPath, "pipeline.json"))
}

func loadFollowUpData(dataPath string) *FollowUpData {
	return loadJSONFile[FollowUpData](filepath.Join(dataPath, "followups.json"))
}

func loadReportsIndexData(dataPath string) *ReportsIndexData {
	return loadJSONFile[ReportsIndexData](filepath.Join(dataPath, "reports-index.json"))
}

func loadStatsData(dataPath string) *Stats {
	return loadJSONFile[Stats](filepath.Join(dataPath, "stats.json"))
}

func loadJSONFile[T any](path string) *T {
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "erro lendo %s: %v\n", path, err)
		return nil
	}
	var result T
	if err := json.Unmarshal(data, &result); err != nil {
		fmt.Fprintf(os.Stderr, "erro parseando %s: %v\n", path, err)
		return nil
	}
	return &result
}

func runBot(cfg *Config) error {
	if err := setupLogging(); err != nil {
		return err
	}

	if cfg.BotToken == "" {
		return fmt.Errorf("BOT_TOKEN não definido. Configure no .env")
	}

	tm := NewTaskManager()

	bot, err := SetupBot(cfg, tm)
	if err != nil {
		return err
	}

	log.Println("🤖 Bot iniciado.")
	bot.Start()
	return nil
}
