package telegram

import (
	"fmt"
	"io"
	"log"
	"os"
)

const logFilePath = "/tmp/telegram-bot.log"

// SetupLogging redireciona o log padrão pra stdout + /tmp/telegram-bot.log,
// pra não precisar redirecionar manualmente ao rodar o bot.
func SetupLogging() error {
	f, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("abrindo %s: %w", logFilePath, err)
	}
	log.SetOutput(io.MultiWriter(os.Stdout, f))
	log.SetFlags(log.Ldate | log.Ltime)
	return nil
}
