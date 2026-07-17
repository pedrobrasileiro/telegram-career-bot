# Telegram Career Bot

Bot Telegram para acessar o career-ops pelo celular.

## Setup

```bash
cd telegram-career-bot
cp .env.example .env
# Editar BOT_TOKEN no .env
```

Crie o bot no [@BotFather](https://t.me/BotFather) e cole o token no `.env`.

`OP_BOT_CAREER_OPS_PATH` no `.env` deve apontar pro clone local de [santifer/career-ops](https://github.com/santifer/career-ops).

## Comandos

```bash
make sync     # Exporta dados do career-ops → data/*.json
make run      # Inicia o bot (long-polling)
make build    # Compila o binário
```

Ou com `go`:

```bash
go run ./cmd/bot sync
go run ./cmd/bot
```

## Comandos do Bot

| Gatilho | Ação |
|---------|------|
| `/start` | Menu de comandos |
| Enviar URL | Auto-pipeline: avalia a vaga |
| `/scan` | Scan completo (scan.mjs + WebSearch L3) |
| `/status` | Job ativo e tempo de execução |
| `/cancel` | Cancela o job em execução (pede confirmação) |
| `/tracker [status]` | Resumo do tracker (filtro opcional) |
| `/stats` | Funil completo (ever applied → offer) |
| `/pipeline` | URLs pendentes |
| `/followup` | Follow-ups próximos 7 dias |
| `/report 194` | Score, veredito, stack de um report |
| `/agenda` | Próximas entrevistas |

## Manutenção

Após avaliar vagas ou rodar scan, atualize os snapshots:

```bash
go run ./cmd/bot sync
```

## Logs

```bash
tail -f /tmp/telegram-bot.log
```

## Estrutura

Clean Architecture: `domain` (entidades) ← `usecase` (regra de negócio, sem I/O) ← `port` (interfaces) ← `adapter` (implementações concretas: Telegram, opencode CLI, markdown, JSON).

```
telegram-career-bot/
  cmd/bot/main.go              # Entry: go run ./cmd/bot → bot, sync → export. Monta os adapters e injeta nos usecases.

  internal/
    config/         # Config (.env + config.json)
    domain/         # Entidades puras: Application, Pipeline, ReportSummary, Stats, Task...
    port/           # Interfaces: JobRunner, CareerOpsSource, SnapshotStore, Notifier
    usecase/        # Regra de negócio: Export, Evaluate, Scan, Ask, TrackerQuery, StatsQuery...
    adapter/
      opencode/     # Implementa JobRunner (spawna `opencode run`)
      markdown/     # Implementa CareerOpsSource (parse markdown → domain)
      jsonstore/    # Implementa SnapshotStore (data/*.json)
      telegram/     # Implementa Notifier + registra rotas telebot + formata HTML

  data/             # Snapshots JSON (gitignored)
```

Ver [CLAUDE.md](CLAUDE.md) pra detalhes de arquitetura e [TEST.md](TEST.md) pra testes.
