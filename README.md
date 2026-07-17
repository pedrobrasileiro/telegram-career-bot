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
go run . sync
go run .
```

## Comandos do Bot

| Gatilho | Ação |
|---------|------|
| `/start` | Menu de comandos |
| Enviar URL | Auto-pipeline: avalia a vaga |
| `/scan` | Scan completo (scan.mjs + WebSearch L3) |
| `/status` | Job ativo e tempo de execução |
| `/tracker [status]` | Resumo do tracker (filtro opcional) |
| `/stats` | Funil completo (ever applied → offer) |
| `/pipeline` | URLs pendentes |
| `/followup` | Follow-ups próximos 7 dias |
| `/report 194` | Score, veredito, stack de um report |
| `/agenda` | Próximas entrevistas |

## Manutenção

Após avaliar vagas ou rodar scan, atualize os snapshots:

```bash
go run . sync
```

## Logs

```bash
tail -f /tmp/telegram-bot.log
```

## Estrutura

```
telegram-career-bot/
  main.go           # Entry: go run . → bot, go run . sync → export
  bot.go            # Setup telebot + data loaders
  handlers.go       # Comandos (start, scan, tracker, stats...)
  parser.go         # Parse markdown → JSON
  export.go         # Exporta dados do career-ops
  opencode.go       # Spawna opencode run
  taskmanager.go    # Jobs em execução (in-memory)
  config.go         # Config (.env + config.json)
  data/             # Snapshots JSON (gitignored)
```
