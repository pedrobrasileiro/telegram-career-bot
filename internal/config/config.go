package config

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	BotToken      string
	EvalTimeout   time.Duration
	ScanTimeout   time.Duration
	CareerOpsPath string
	DataPath      string
	Language      string
}

type configFile struct {
	CareerOpsPath string `json:"careerOpsPath"`
	DataPath      string `json:"dataPath"`
	Language      string `json:"language"`
}

func Load() (*Config, error) {
	loadDotEnv()

	cfg := &Config{
		BotToken:      os.Getenv("BOT_TOKEN"),
		EvalTimeout:   300 * time.Second,
		ScanTimeout:   600 * time.Second,
		CareerOpsPath: "../career-ops",
		DataPath:      "./data",
		Language:      "pt-BR",
	}

	if v := os.Getenv("OP_BOT_EVAL_TIMEOUT_MS"); v != "" {
		if ms, err := strconv.Atoi(v); err == nil && ms > 0 {
			cfg.EvalTimeout = time.Duration(ms) * time.Millisecond
		}
	}
	if v := os.Getenv("OP_BOT_SCAN_TIMEOUT_MS"); v != "" {
		if ms, err := strconv.Atoi(v); err == nil && ms > 0 {
			cfg.ScanTimeout = time.Duration(ms) * time.Millisecond
		}
	}
	if v := os.Getenv("OP_BOT_CAREER_OPS_PATH"); v != "" {
		cfg.CareerOpsPath = v
	}

	if data, err := os.ReadFile("config.json"); err == nil {
		var f configFile
		if json.Unmarshal(data, &f) == nil {
			if cfg.CareerOpsPath == "../career-ops" && f.CareerOpsPath != "" {
				cfg.CareerOpsPath = f.CareerOpsPath
			}
			if f.DataPath != "" {
				cfg.DataPath = f.DataPath
			}
			if f.Language != "" {
				cfg.Language = f.Language
			}
		}
	}

	return cfg, nil
}

func loadDotEnv() {
	f, err := os.Open(".env")
	if err != nil {
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		idx := strings.IndexByte(line, '=')
		if idx < 0 {
			continue
		}
		key := strings.TrimSpace(line[:idx])
		val := strings.TrimSpace(line[idx+1:])
		if os.Getenv(key) == "" {
			os.Setenv(key, val)
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "erro lendo .env: %v\n", err)
	}
}
