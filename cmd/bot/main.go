package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"telegram-career-bot/internal/adapter/jsonstore"
	"telegram-career-bot/internal/adapter/markdown"
	"telegram-career-bot/internal/adapter/opencode"
	"telegram-career-bot/internal/adapter/telegram"
	"telegram-career-bot/internal/config"
	"telegram-career-bot/internal/usecase"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "sync" {
		runSync()
		return
	}

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "erro carregando config: %v\n", err)
		os.Exit(1)
	}

	if err := runBot(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "erro no bot: %v\n", err)
		os.Exit(1)
	}
}

func runSync() {
	careerOpsPath := "../career-ops"
	dataPath := "./data"

	cfg, err := config.Load()
	if err == nil {
		careerOpsPath = cfg.CareerOpsPath
		dataPath = cfg.DataPath
	}

	fmt.Printf("Exportando dados de %s para %s...\n", careerOpsPath, dataPath)

	exporter := usecase.Export{
		Source: markdown.New(careerOpsPath),
		Store:  jsonstore.New(dataPath),
	}
	if err := exporter.Run(); err != nil {
		fmt.Printf("Erro: %v\n", err)
		return
	}

	fmt.Println("Export concluído:")
	fmt.Printf("  tracker.json\n  pipeline.json\n  followups.json\n  reports-index.json\n  stats.json\n")
}

func runBot(cfg *config.Config) error {
	if err := telegram.SetupLogging(); err != nil {
		return err
	}

	if cfg.BotToken == "" {
		return fmt.Errorf("BOT_TOKEN não definido. Configure no .env")
	}

	store := jsonstore.New(cfg.DataPath)
	source := markdown.New(cfg.CareerOpsPath)
	runner := opencode.New()
	exporter := usecase.Export{Source: source, Store: store}

	deps := telegram.Deps{
		BotToken:      cfg.BotToken,
		CareerOpsPath: cfg.CareerOpsPath,
		Tracker:       usecase.TrackerQuery{Store: store},
		Stats:         usecase.StatsQuery{Store: store},
		Pipeline:      usecase.PipelineQuery{Store: store},
		FollowUp:      usecase.FollowUpQuery{Store: store},
		Report:        usecase.ReportQuery{Store: store},
		Agenda:        usecase.AgendaQuery{Store: store},
		Evaluate: usecase.Evaluate{
			Runner:        runner,
			Exporter:      exporter,
			Store:         store,
			CareerOpsPath: cfg.CareerOpsPath,
			Timeout:       cfg.EvalTimeout,
		},
		Scan: usecase.Scan{
			Runner:        runner,
			Exporter:      exporter,
			Store:         store,
			CareerOpsPath: cfg.CareerOpsPath,
			Timeout:       cfg.ScanTimeout,
		},
		Ask: usecase.Ask{
			Runner:        runner,
			Store:         store,
			CareerOpsPath: cfg.CareerOpsPath,
			Timeout:       120 * time.Second,
		},
	}

	bot, err := telegram.SetupBot(deps)
	if err != nil {
		return err
	}

	log.Println("🤖 Bot iniciado.")
	bot.Start()
	return nil
}
