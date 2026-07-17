# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## O que é

Bot Telegram (Go, `gopkg.in/telebot.v3`) que dá acesso móvel ao projeto irmão `career-ops` (repo separado, path configurável, padrão `../career-ops`). O bot não contém lógica de avaliação de vagas — ele lê/escreve arquivos markdown do career-ops, dispara `opencode run` lá dentro, e expõe os resultados via comandos Telegram.

## Comandos

```bash
make sync     # go run . sync — exporta markdown do career-ops → data/*.json
make run      # go run .      — inicia bot em long-polling
make build    # compila binário ./telegram-career-bot
```

Não há testes automatizados nem linter configurado neste repo.

## Arquitetura

Fluxo de dados é unidirecional e baseado em snapshot, não em leitura direta:

```
career-ops/data/*.md, career-ops/reports/*.md
        │  parser.go (parseApplications, parsePipeline, parseFollowUps, parseReportHeaders)
        ▼
export.go: exportAll()
        │  escreve
        ▼
data/*.json (tracker, pipeline, followups, reports-index, stats)  ← gitignored
        │  bot.go: loadTrackerData / loadPipelineData / loadFollowUpData / loadReportsIndexData / loadStatsData
        ▼
handlers.go: /tracker, /stats, /pipeline, /followup, /report, /agenda
```

Os handlers **nunca** leem o career-ops diretamente — só os JSONs em `data/`. Por isso, sempre que uma avaliação ou scan roda, é preciso re-exportar (`exportAll`) antes que os comandos reflitam o estado novo. `handleScan` e `handleEval` já chamam `exportAll` automaticamente ao final; se os dados parecerem desatualizados, rode `go run . sync` manualmente.

Se `data/*.json` não existir, os handlers retornam `handleSyncNeeded` (mensagem pedindo `go run . sync`) em vez de crashar.

### Divisão de arquivos (não corresponde 1:1 ao README)

- `main.go` — entrypoint: `sync` como subcomando ou inicia o bot.
- `bot.go` — `runBot`, loaders de JSON (`loadJSONFile[T]` genérico), helpers (`isURL`, `parseInt`, `filepathFromCareerOps`).
- `handlers.go` — `SetupBot` (registro de todas as rotas do telebot) + implementação de cada handler.
- `config.go` — `Config`, leitura de `.env` (parser próprio, sem lib externa) e `config.json`; env vars têm prioridade sobre `config.json`.
- `parser.go` — parsing de markdown do career-ops (tabelas `| # | ... |`, checkboxes `- [ ]`) para as structs `Application`, `PipelineItem`, `FollowUpItem`, `ReportSummary`; e `computeStats`.
- `export.go` — orquestra os parsers e escreve os 5 JSONs em `data/`.
- `opencode.go` — `runOpenCode`: roda `opencode run <prompt>` via `exec.CommandContext` com `cmd.Dir = careerOpsPath`, timeout configurável.
- `taskmanager.go` — `TaskManager`: mapa `chatID → *Task` em memória (protegido por mutex) para impedir jobs concorrentes por chat (`/scan`, avaliação de URL, `/ask` só um por vez).

### Handlers assíncronos (opencode)

`/scan`, envio de URL (auto-avaliação) e perguntas em linguagem natural (`handleAsk`) disparam `runOpenCode` numa goroutine, respondem imediatamente com "iniciado", e enviam o resultado depois via `bot.Send(c.Sender(), ...)`. `TaskManager.IsBusy` bloqueia novo job enquanto um está rodando pro mesmo chat; `/status` consulta o job ativo.

### Config

Precedência: env vars (`.env` ou ambiente) > `config.json` > defaults hardcoded em `loadConfig()`. `OP_BOT_CAREER_OPS_PATH`, `OP_BOT_EVAL_TIMEOUT_MS`, `OP_BOT_SCAN_TIMEOUT_MS` são as únicas env vars customizadas lidas (além de `BOT_TOKEN`).

### Diretório `handlers/`

Existe mas está vazio — todo o código de handlers vive em `handlers.go` na raiz.
