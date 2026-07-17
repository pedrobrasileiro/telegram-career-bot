package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "sync" {
		runSync()
		return
	}

	cfg, err := loadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "erro carregando config: %v\n", err)
		os.Exit(1)
	}

	if err := runBot(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "erro no bot: %v\n", err)
		os.Exit(1)
	}
}
