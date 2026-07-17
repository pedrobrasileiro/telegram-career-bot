package telegram

import (
	"fmt"
	"log"
	"strings"
	"time"

	"gopkg.in/telebot.v3"

	"telegram-career-bot/internal/port"
	"telegram-career-bot/internal/usecase"
)

// Deps agrupa tudo que o SetupBot precisa: token e os usecases já
// montados com seus adapters concretos (opencode, markdown, jsonstore).
type Deps struct {
	BotToken      string
	CareerOpsPath string

	Tracker  usecase.TrackerQuery
	Stats    usecase.StatsQuery
	Pipeline usecase.PipelineQuery
	FollowUp usecase.FollowUpQuery
	Report   usecase.ReportQuery
	Agenda   usecase.AgendaQuery
	Evaluate usecase.Evaluate
	Scan     usecase.Scan
	Ask      usecase.Ask
}

func SetupBot(deps Deps) (*telebot.Bot, error) {
	pref := telebot.Settings{
		Token:  deps.BotToken,
		Poller: &telebot.LongPoller{Timeout: 10 * time.Second},
	}

	bot, err := telebot.NewBot(pref)
	if err != nil {
		return nil, fmt.Errorf("criando bot: %w", err)
	}

	var notifier port.Notifier = Notifier{Bot: bot}
	tm := NewTaskManager()
	ai := NewAwaitingInput()

	bot.Use(func(next telebot.HandlerFunc) telebot.HandlerFunc {
		return func(c telebot.Context) error {
			sender := c.Sender()
			log.Printf("comando de %s (id=%d): %q", sender.Username, sender.ID, c.Text())
			return next(c)
		}
	})

	bot.Handle("/start", func(c telebot.Context) error {
		return c.Send(renderHelp(), &telebot.SendOptions{ParseMode: telebot.ModeHTML})
	})

	bot.Handle("/help", func(c telebot.Context) error {
		return c.Send(renderHelp(), &telebot.SendOptions{ParseMode: telebot.ModeHTML})
	})

	bot.Handle("/scan", func(c telebot.Context) error {
		return handleScan(c, tm, deps.Scan, notifier)
	})

	bot.Handle("/status", func(c telebot.Context) error {
		return handleStatus(c, tm)
	})

	bot.Handle("/vstatus", func(c telebot.Context) error {
		return handleStatus(c, tm)
	})

	bot.Handle("/tracker", func(c telebot.Context) error {
		filter := strings.TrimSpace(strings.TrimPrefix(c.Text(), "/tracker"))
		if filter == "" {
			ai.Set(c.Sender().ID, "tracker")
			return c.Send("Digite o status pra filtrar (ex: Interview) ou \"todos\":", &telebot.ReplyMarkup{ForceReply: true})
		}
		return handleTrackerFiltered(c, deps.Tracker, filter)
	})

	bot.Handle("/stats", func(c telebot.Context) error {
		return handleStats(c, deps.Stats)
	})

	bot.Handle("/pipeline", func(c telebot.Context) error {
		return handlePipeline(c, deps.Pipeline)
	})

	bot.Handle("/followup", func(c telebot.Context) error {
		return handleFollowUp(c, deps.FollowUp)
	})

	bot.Handle("/report", func(c telebot.Context) error {
		args := strings.Fields(c.Text())
		if len(args) < 2 {
			ai.Set(c.Sender().ID, "report")
			return c.Send("Digite o número do report (ex: 194):", &telebot.ReplyMarkup{ForceReply: true})
		}
		return handleReport(c, deps.Report, deps.CareerOpsPath)
	})

	bot.Handle("/agenda", func(c telebot.Context) error {
		return handleAgenda(c, deps.Agenda)
	})

	bot.Handle("/cancel", func(c telebot.Context) error {
		return handleCancel(c, tm)
	})

	bot.Handle(&telebot.Btn{Unique: btnCancelYes}, func(c telebot.Context) error {
		return handleCancelConfirm(c, tm)
	})

	bot.Handle(&telebot.Btn{Unique: btnCancelNo}, handleCancelAbort)

	bot.Handle(telebot.OnText, func(c telebot.Context) error {
		text := c.Text()
		chatID := c.Sender().ID

		if kind, ok := ai.Take(chatID); ok {
			switch kind {
			case "report":
				num, err := parseInt(text)
				if err != nil {
					return c.Send("Número inválido.")
				}
				return handleReportByNum(c, deps.Report, num, deps.CareerOpsPath)
			case "tracker":
				filter := strings.TrimSpace(text)
				if strings.EqualFold(filter, "todos") {
					filter = ""
				}
				return handleTrackerFiltered(c, deps.Tracker, filter)
			}
		}

		if isURL(text) {
			return handleEval(c, tm, deps.Evaluate, notifier, text)
		}
		if strings.HasPrefix(text, "/") {
			return nil // unknown command, ignore
		}
		return handleAsk(c, tm, deps.Ask, notifier, text)
	})

	if err := bot.SetCommands([]telebot.Command{
		{Text: "scan", Description: "Scan completo de vagas"},
		{Text: "status", Description: "Job ativo em execução"},
		{Text: "cancel", Description: "Cancela o job em execução"},
		{Text: "tracker", Description: "Resumo do tracker"},
		{Text: "stats", Description: "Funil completo"},
		{Text: "pipeline", Description: "URLs pendentes"},
		{Text: "followup", Description: "Follow-ups próximos 7 dias"},
		{Text: "report", Description: "Resumo de um report (ex: /report 194)"},
		{Text: "agenda", Description: "Próximas entrevistas"},
		{Text: "help", Description: "Menu de comandos"},
	}); err != nil {
		log.Printf("erro registrando menu de comandos: %v", err)
	}

	return bot, nil
}

func renderHelp() string {
	return `<b>📋 Telegram Career Bot</b>

Comandos:

<b>Enviar URL</b> — Avalia a vaga (auto-pipeline)
<b>/scan</b> — Scan completo de vagas
<b>/status</b> — Job ativo em execução
<b>/cancel</b> — Cancela o job em execução (com confirmação)
<b>/tracker [status]</b> — Resumo do tracker
<b>/stats</b> — Funil completo
<b>/pipeline</b> — URLs pendentes
<b>/followup</b> — Follow-ups próximos 7 dias
<b>/report 194</b> — Resumo de um report
<b>/agenda</b> — Próximas entrevistas`
}
