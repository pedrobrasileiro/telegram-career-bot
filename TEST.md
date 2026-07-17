# Testes

## Libs

Escrita dos testes: nenhuma lib externa. Só `testing` (stdlib) — sem testify, sem mocks, sem framework de asserção. Padrão Go puro: `t.Errorf`/`t.Fatalf` + comparação manual, `t.TempDir()` pra fixtures em disco.

Execução/relatório: [gotestsum](https://github.com/gotestyourself/gotestsum) — CLI que roda `go test` por baixo e formata a saída (tipo Jest: um `PASS`/`FAIL` por teste + resumo `DONE N tests` no final). Não é dependência do módulo (não entra no `go.mod`), é uma ferramenta de dev.

Instalar (uma vez):

```bash
CGO_ENABLED=0 go install gotest.tools/gotestsum@latest
```

## Rodar

```bash
make test
```

Equivale a `CGO_ENABLED=0 gotestsum --format testname ./...`. `CGO_ENABLED=0` necessário no ambiente atual (linker Go 1.21.13 + Xcode dá erro `missing LC_UUID load command` com CGO ligado — mesma causa do build normal).

Sem `gotestsum` instalado, cai pro padrão:

```bash
CGO_ENABLED=0 go test ./... -v              # verbose, lista cada teste
CGO_ENABLED=0 go test ./... -run TestIsURL  # roda só um teste (regex no nome)
CGO_ENABLED=0 go test ./... -cover          # % de cobertura
```

## O que é coberto

Só funções puras (sem I/O de rede, sem Telegram, sem `opencode`). Cada teste mora junto do arquivo/camada que testa:

| Pacote | Arquivo | Testa |
|---|---|---|
| `internal/adapter/markdown` | `source_test.go` | Parsing de markdown do career-ops (`parseTableRow`, `cleanStatus`, `extractReportPath`, `parseNotes`, `cleanFollowUpDate`) e os métodos completos de `Source` (`ParseApplications`, `ParsePipeline`, `ParseFollowUps`, `ParseReports`) usando fixtures escritas em `t.TempDir()`. |
| `internal/adapter/jsonstore` | `store_test.go` | `writeJSON`/`loadJSON` round-trip e arquivo ausente. |
| `internal/adapter/opencode` | `runner_test.go` | `lastNonEmptyLine`, `logWriter` (acúmulo + escrita fragmentada, streaming linha a linha pro log). |
| `internal/adapter/telegram` | `util_test.go`, `taskmanager_test.go`, `handlers_test.go` | `isURL`, `parseInt`, `filepathFromCareerOps`; ciclo de vida do `TaskManager` (Start/IsBusy/Get/End, isolamento entre chats) e `AwaitingInput` (Set/Take consumindo a pendência); `emojiForTrackerStatus`, `emojiForTask` (presentation). |
| `internal/usecase` | `stats_test.go`, `scan_test.go`, `evaluate_test.go` | `ComputeStats`, `StatusOrder`, `extractScanSummary`, `lastNonEmptyLine`. |

## O que NÃO é coberto (e por quê)

- **`adapter/telegram` handlers que recebem `telebot.Context`** (`handleTracker`, `handleReport`, `handleStats`, etc.) — dependem de `telebot.Context`, que não é trivial de mockar sem lib extra. Testável indiretamente rodando o bot de verdade. A lógica de negócio por trás deles (filtro, agregação, formatação de dados) já está isolada e testada em `internal/usecase`.
- **`adapter/opencode.Runner.Run` / `adapter/telegram.SetupBot`** — dependem do binário externo `opencode` e da API do Telegram (rede). Fora do escopo de teste unitário.
- **`internal/config.Load`** — lê `.env`/`config.json` do diretório atual; não isolado o suficiente pra testar sem side effects (mexe em env vars globais via `os.Setenv`).

Como os `port.*` (JobRunner, CareerOpsSource, SnapshotStore, Notifier) já são interfaces, dá pra escrever fakes simples e testar os usecases (`Evaluate`, `Scan`, `Ask`) em isolamento sem precisar de nenhuma lib de mock — ainda não feito porque a cobertura atual (parsing + regra de agregação) já protege a parte mais frágil (parsing de markdown solto).
