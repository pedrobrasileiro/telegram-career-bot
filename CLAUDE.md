# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## O que é

Bot Telegram (Go, `gopkg.in/telebot.v3`) que dá acesso móvel ao projeto irmão `career-ops` (repo separado, path configurável, padrão `../career-ops`). O bot não contém lógica de avaliação de vagas — ele lê/escreve arquivos markdown do career-ops, dispara `opencode run` lá dentro, e expõe os resultados via comandos Telegram.

## Comandos

```bash
make sync     # go run ./cmd/bot sync — exporta markdown do career-ops → data/*.json
make run      # go run ./cmd/bot      — inicia bot em long-polling
make build    # compila binário ./telegram-career-bot
make test     # gotestsum --format testname ./... — relatório por teste + resumo
```

Ver [TEST.md](TEST.md) pra detalhes de teste.

## Arquitetura

Clean Architecture, camadas de dentro pra fora: `domain` (entidades) ← `usecase` (regra de negócio, zero I/O, zero telebot) ← `port` (interfaces/contratos) ← `adapter` (implementações concretas). `cmd/bot/main.go` é o único lugar que conhece todos os adapters concretos e faz a injeção de dependência manual (sem framework de DI).

```
internal/
  domain/    Application, Pipeline, PipelineItem, FollowUpItem, ReportSummary,
             Stats, Funnel, Task — structs puras, sem métodos de I/O.

  port/      Interfaces que usecase depende (Dependency Inversion):
               JobRunner        — roda um prompt no career-ops, retorna output (aceita
                                  context.Context, cancelável via /cancel)
               CareerOpsSource  — parseia o markdown do career-ops → domain
               SnapshotStore    — lê/escreve o snapshot data/*.json
               Notifier         — envia mensagem assíncrona pra um chat

  usecase/   Regra de negócio pura (nunca importa telebot, nunca monta HTML):
               Export                              — parseia career-ops → escreve snapshot
               Evaluate, Scan, Ask                  — jobs que rodam opencode + reexportam
               TrackerQuery, StatsQuery, PipelineQuery, FollowUpQuery,
               ReportQuery, AgendaQuery              — leitura/filtro/agregação sobre o snapshot
               ComputeStats, StatusOrder             — funções auxiliares puras

  adapter/
    opencode/   Runner: implementa JobRunner via exec.CommandContext("opencode", "run", prompt).
    markdown/   Source: implementa CareerOpsSource (parse de tabelas `| # | ... |` e checkboxes `- [ ]`).
    jsonstore/  Store: implementa SnapshotStore (data/*.json).
    telegram/   Notifier (implementa port.Notifier) + SetupBot (registra rotas telebot,
                formata HTML/emoji, TaskManager/AwaitingInput — estado de sessão por chat).

internal/config/  Config: leitura de .env (parser próprio) + config.json.
cmd/bot/main.go   Entry point: `sync` como subcomando ou inicia o bot; monta os adapters
                  concretos e injeta em cada usecase antes de passar pro telegram.SetupBot.
```

### Fluxo de dados (snapshot, não leitura direta)

```
career-ops/data/*.md, career-ops/reports/*.md
        │  adapter/markdown.Source (implementa port.CareerOpsSource)
        ▼
usecase.Export.Run()
        │  escreve via adapter/jsonstore.Store (port.SnapshotStore)
        ▼
data/*.json (tracker, pipeline, followups, reports-index, stats)  ← gitignored
        │  usecase.TrackerQuery / StatsQuery / PipelineQuery / FollowUpQuery / ReportQuery / AgendaQuery
        ▼
adapter/telegram: /tracker, /stats, /pipeline, /followup, /report, /agenda
```

Os usecases de leitura **nunca** leem o career-ops diretamente — só o snapshot em `data/`. Por isso, sempre que uma avaliação ou scan roda, é preciso reexportar antes que os comandos reflitam o estado novo. `usecase.Evaluate.Run` e `usecase.Scan.Run` já chamam `Export.Run()` internamente ao final (erro de export não bloqueia a resposta, mesmo comportamento de antes); se os dados parecerem desatualizados, rode `go run ./cmd/bot sync` manualmente.

Se o snapshot não existir, os usecases de query retornam erro e `adapter/telegram` responde com `handleSyncNeeded` (mensagem pedindo `go run ./cmd/bot sync`) em vez de crashar.

### Regra de ouro entre camadas

- `usecase/` nunca importa `gopkg.in/telebot.v3` nem monta string HTML — só `domain` + `port`. Retorna structs de resultado (`TrackerResult`, `StatsResult`, `EvaluateResult`...).
- `adapter/telegram/handlers.go` é o único lugar que formata `<b>...</b>`, escolhe emoji (`emojiForTrackerStatus`, `emojiForTask`) e monta `telebot.SendOptions`.
- Jobs assíncronos (`/scan`, envio de URL, `/ask`) continuam disparados por `adapter/telegram` (é lá que `telebot.Context`/`TaskManager` vivem — máximo um job concorrente por chat via `TaskManager.IsBusy`), mas o trabalho pesado (rodar opencode, reexportar, buscar resultado) é feito pelo usecase (`usecase.Scan`, `usecase.Evaluate`, `usecase.Ask`); o handler só formata o retorno e envia via `port.Notifier`.

### Cancelamento de job (`/cancel`)

Cada goroutine de job (`handleScan`/`handleEval`/`handleAsk` em `adapter/telegram/handlers.go`) cria um `context.WithCancel` e guarda o `cancel func` no `TaskManager` via `Start(chatID, type, desc, cancel)`. `/cancel` pede confirmação por botão inline (`handleCancel` → `cancel_yes`/`cancel_no`); confirmando, `TaskManager.Cancel(chatID)` dispara o cancel func, que propaga pro `context.Context` passado a `usecase.Scan.Run(ctx)`/`Evaluate.Run(ctx, url)`/`Ask.Run(ctx, question)` → `port.JobRunner.Run(ctx, ...)` → `exec.CommandContext` mata o processo `opencode`. O handler detecta `errors.Is(err, context.Canceled)` pra mandar "🛑 ... cancelado" em vez do erro genérico.

### Config

Precedência: env vars (`.env` ou ambiente) > `config.json` > defaults hardcoded em `config.Load()`. `OP_BOT_CAREER_OPS_PATH`, `OP_BOT_EVAL_TIMEOUT_MS`, `OP_BOT_SCAN_TIMEOUT_MS` são as únicas env vars customizadas lidas (além de `BOT_TOKEN`).

### Logs

`adapter/telegram.SetupLogging()` (chamado por `cmd/bot/main.go` no boot do bot) redireciona `log.Printf` pra stdout + `/tmp/telegram-bot.log` — não precisa redirecionar manualmente ao rodar. `adapter/opencode.Runner` loga comando executado e retorno completo do `opencode run`.
