package port

import (
	"context"
	"time"
)

// JobRunner executa um comando de longa duração (o CLI `opencode`) num
// diretório de trabalho e retorna a saída combinada (stdout+stderr).
// O ctx permite cancelamento externo (ex: usuário pede /cancel) além do
// timeout.
type JobRunner interface {
	Run(ctx context.Context, workDir, prompt string, timeout time.Duration) (string, error)
}
