package opencode

import (
	"bytes"
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

	lw := &logWriter{}
	cmd.Stdout = lw
	cmd.Stderr = lw

	log.Printf("opencode: executando em %s: opencode run %q (timeout %v)", workDir, prompt, timeout)

	err := cmd.Run()
	lw.flush()
	out := lw.buf.Bytes()

	log.Printf("opencode: concluído (%d bytes de saída)", len(out))

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

// logWriter acumula tudo que o subprocesso escreve (pro retorno da função)
// e ao mesmo tempo loga cada linha assim que ela chega — pra dar
// visibilidade em tempo real de um job longo (ex: /scan levando minutos)
// em vez de só logar tudo de uma vez quando o processo termina.
type logWriter struct {
	buf     bytes.Buffer
	lineBuf []byte
}

func (w *logWriter) Write(p []byte) (int, error) {
	w.buf.Write(p)
	w.lineBuf = append(w.lineBuf, p...)

	for {
		idx := bytes.IndexByte(w.lineBuf, '\n')
		if idx < 0 {
			break
		}
		line := strings.TrimRight(string(w.lineBuf[:idx]), "\r")
		if strings.TrimSpace(line) != "" {
			log.Printf("opencode: %s", line)
		}
		w.lineBuf = w.lineBuf[idx+1:]
	}

	return len(p), nil
}

// flush loga qualquer resto sem newline final (última linha de saída
// sem quebra de linha no fim).
func (w *logWriter) flush() {
	if line := strings.TrimSpace(string(w.lineBuf)); line != "" {
		log.Printf("opencode: %s", line)
	}
	w.lineBuf = nil
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
