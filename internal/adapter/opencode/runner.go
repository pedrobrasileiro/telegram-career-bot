package opencode

import (
	"context"
	"fmt"
	"log"
	"os/exec"
	"strings"
	"time"
)

// Runner implementa port.JobRunner rodando o CLI `opencode` como
// subprocesso.
type Runner struct{}

func New() Runner { return Runner{} }

func (Runner) Run(ctx context.Context, workDir, prompt string, timeout time.Duration) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "opencode", "run", prompt)
	cmd.Dir = workDir

	log.Printf("opencode: executando em %s: opencode run %q (timeout %v)", workDir, prompt, timeout)

	out, err := cmd.CombinedOutput()
	log.Printf("opencode: retorno (%d bytes):\n%s", len(out), string(out))

	if ctx.Err() == context.Canceled {
		log.Printf("opencode: cancelado pelo usuário")
		return "", context.Canceled
	}
	if ctx.Err() == context.DeadlineExceeded {
		log.Printf("opencode: timeout após %v", timeout.Round(time.Second))
		return "", fmt.Errorf("timeout após %v", timeout.Round(time.Second))
	}
	if err != nil {
		log.Printf("opencode: erro: %v", err)
		lastLine := lastNonEmptyLine(string(out))
		if lastLine != "" {
			return "", fmt.Errorf("erro: %s", lastLine)
		}
		return "", fmt.Errorf("erro (código %v)", err)
	}

	return string(out), nil
}

func lastNonEmptyLine(s string) string {
	lines := strings.Split(strings.TrimSpace(s), "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		if line := strings.TrimSpace(lines[i]); line != "" {
			return line
		}
	}
	return ""
}
