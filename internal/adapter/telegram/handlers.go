package telegram

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"gopkg.in/telebot.v3"

	"telegram-career-bot/internal/port"
	"telegram-career-bot/internal/usecase"
)

func handleStatus(c telebot.Context, tm *TaskManager) error {
	task := tm.Get(c.Sender().ID)
	if task == nil {
		return c.Send("✅ Nenhum job rodando.")
	}
	elapsed := time.Since(task.StartTime).Round(time.Second)
	return c.Send(fmt.Sprintf("🔍 <b>%s</b> — %s\n⏱ Rodando há %v\n\n<b>Descrição:</b> %s",
		emojiForTask(task.Type), task.Type, elapsed, task.Description),
		&telebot.SendOptions{ParseMode: telebot.ModeHTML})
}

func emojiForTask(taskType string) string {
	switch taskType {
	case "scan":
		return "🔎"
	case "avaliação":
		return "📊"
	default:
		return "⏳"
	}
}

func handleTrackerFiltered(c telebot.Context, tracker usecase.TrackerQuery, filter string) error {
	result, err := tracker.Run(filter)
	if err != nil {
		return handleSyncNeeded(c)
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("<b>📊 Tracker</b> (%d vagas)\n\n", result.Total))
	sb.WriteString("<b>Por status:</b>\n")
	for _, s := range usecase.StatusOrder() {
		if count, ok := result.StatusCounts[s]; ok && count > 0 {
			sb.WriteString(fmt.Sprintf("  %s %s: %d\n", emojiForTrackerStatus(s), s, count))
		}
	}

	if filter != "" {
		sb.WriteString(fmt.Sprintf("\n<b>Filtro: %s</b>\n", filter))
	}

	limit := 5
	if len(result.Filtered) < limit {
		limit = len(result.Filtered)
	}
	sb.WriteString(fmt.Sprintf("\n<b>Últimas %d:</b>\n", limit))
	for i := 0; i < limit; i++ {
		a := result.Filtered[i]
		emoji := emojiForTrackerStatus(a.Status)
		sb.WriteString(fmt.Sprintf("%s <b>#%d</b> %s — %s (%s)\n", emoji, a.Num, a.Company, a.Role, a.Status))
	}

	if len(result.Filtered) > limit {
		sb.WriteString(fmt.Sprintf("\n... +%d mais\n", len(result.Filtered)-limit))
	}

	return c.Send(sb.String(), &telebot.SendOptions{ParseMode: telebot.ModeHTML})
}

func emojiForTrackerStatus(s string) string {
	switch s {
	case "Applied":
		return "📤"
	case "Interview":
		return "🎯"
	case "Offer":
		return "🎉"
	case "Rejected":
		return "❌"
	case "Discarded":
		return "🗑"
	case "Evaluated":
		return "📋"
	case "SKIP":
		return "⏭"
	default:
		return "📌"
	}
}

func handleStats(c telebot.Context, stats usecase.StatsQuery) error {
	result, err := stats.Run()
	if err != nil {
		return handleSyncNeeded(c)
	}

	var sb strings.Builder
	sb.WriteString("<b>📈 Stats</b>\n\n")
	sb.WriteString(fmt.Sprintf("<b>Total:</b> %d | <b>Ativas:</b> %d\n", result.Stats.Total, result.Active))
	sb.WriteString(fmt.Sprintf("<b>Avg fit:</b> %.1f/5\n\n", result.Stats.AverageScore))

	sb.WriteString("<b>Por status:</b>\n")
	for _, s := range usecase.StatusOrder() {
		if c, ok := result.Stats.ByStatus[s]; ok && c > 0 {
			sb.WriteString(fmt.Sprintf("  %s %s: %d\n", emojiForTrackerStatus(s), s, c))
		}
	}

	sb.WriteString("\n<b>Funil:</b>\n")
	sb.WriteString(fmt.Sprintf("  📤 Aplicadas: %d\n", result.EverApplied))
	sb.WriteString(fmt.Sprintf("  📩 Responderam: %d (%.0f%%)\n", result.EverResponded, result.ResponseRate))
	sb.WriteString(fmt.Sprintf("  🎯 Entrevista: %d (%.0f%%)\n", result.Interview, result.InterviewRate))
	sb.WriteString(fmt.Sprintf("  🎉 Oferta: %d\n", result.Offer))
	sb.WriteString(fmt.Sprintf("  ❌ Rejeitadas: %d\n", result.Rejected))
	sb.WriteString(fmt.Sprintf("  🗑 Descartadas/Skip: %d\n", result.Skipped))

	return c.Send(sb.String(), &telebot.SendOptions{ParseMode: telebot.ModeHTML})
}

func handlePipeline(c telebot.Context, pipelineQuery usecase.PipelineQuery) error {
	pipeline, err := pipelineQuery.Run()
	if err != nil {
		return handleSyncNeeded(c)
	}

	var sb strings.Builder
	sb.WriteString("<b>📥 Pipeline</b>\n\n")

	sb.WriteString(fmt.Sprintf("<b>Pendentes:</b> %d\n", len(pipeline.Pending)))
	for i, item := range pipeline.Pending {
		if i >= 10 {
			sb.WriteString(fmt.Sprintf("\n... +%d mais\n", len(pipeline.Pending)-10))
			break
		}
		sb.WriteString(fmt.Sprintf("  • %s | %s\n", item.Company, item.Title))
	}

	if len(pipeline.Pending) == 0 {
		sb.WriteString("  (vazio)\n")
	}

	return c.Send(sb.String(), &telebot.SendOptions{ParseMode: telebot.ModeHTML})
}

func handleFollowUp(c telebot.Context, followUpQuery usecase.FollowUpQuery) error {
	result, err := followUpQuery.Run(time.Now())
	if err != nil {
		return handleSyncNeeded(c)
	}

	var sb strings.Builder
	sb.WriteString("<b>📅 Follow-ups — próximos 7 dias</b>\n\n")

	if len(result.Upcoming) == 0 {
		sb.WriteString("Nenhum follow-up nos próximos 7 dias.")
	} else {
		for _, item := range result.Upcoming {
			sb.WriteString(fmt.Sprintf("<b>%s</b> — %s\n", item.Date, item.Company))
			if item.Action != "" {
				sb.WriteString(fmt.Sprintf("  Ação: %s\n", item.Action))
			}
		}
	}

	return c.Send(sb.String(), &telebot.SendOptions{ParseMode: telebot.ModeHTML})
}

func handleReport(c telebot.Context, reportQuery usecase.ReportQuery, careerOpsPath string) error {
	args := strings.Fields(c.Text())
	if len(args) < 2 {
		return c.Send("Uso: /report <número>\nEx: /report 194")
	}

	num, err := parseInt(args[1])
	if err != nil {
		return c.Send("Número inválido.")
	}

	return handleReportByNum(c, reportQuery, num, careerOpsPath)
}

func handleReportByNum(c telebot.Context, reportQuery usecase.ReportQuery, num int, careerOpsPath string) error {
	report, err := reportQuery.Run(num)
	if err != nil {
		return handleSyncNeeded(c)
	}
	if report == nil {
		return c.Send(fmt.Sprintf("Report #%d não encontrado.", num))
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("<b>📄 Report #%d</b>\n\n", report.Num))
	sb.WriteString(fmt.Sprintf("<b>Empresa:</b> %s\n", report.Company))
	sb.WriteString(fmt.Sprintf("<b>Cargo:</b> %s\n", report.Role))
	sb.WriteString(fmt.Sprintf("<b>Score:</b> %s\n", report.Score))
	sb.WriteString(fmt.Sprintf("<b>Veredito:</b> %s\n", report.Veredito))

	if report.URL != "" {
		sb.WriteString(fmt.Sprintf("<b>URL:</b> %s\n", report.URL))
	}
	if report.Archetype != "" {
		sb.WriteString(fmt.Sprintf("<b>Arquétipo:</b> %s\n", report.Archetype))
	}
	if report.Legitimacy != "" {
		sb.WriteString(fmt.Sprintf("<b>Legitimidade:</b> %s\n", report.Legitimacy))
	}

	reportsDir := filepathFromCareerOps(careerOpsPath, "reports")
	sb.WriteString(fmt.Sprintf("\n<b>Arquivo:</b> %s/%s", reportsDir, report.Filename))

	return c.Send(sb.String(), &telebot.SendOptions{ParseMode: telebot.ModeHTML})
}

func handleAgenda(c telebot.Context, agendaQuery usecase.AgendaQuery) error {
	interviews, err := agendaQuery.Run()
	if err != nil {
		return handleSyncNeeded(c)
	}

	var sb strings.Builder
	sb.WriteString("<b>🎯 Agenda de Entrevistas</b>\n\n")

	if len(interviews) == 0 {
		sb.WriteString("Nenhuma entrevista agendada no momento.")
	} else {
		for _, a := range interviews {
			sb.WriteString(fmt.Sprintf("<b>#%d</b> %s — %s\n", a.Num, a.Company, a.Role))
			if a.Notes != "" {
				sb.WriteString(fmt.Sprintf("  %s\n", a.Notes))
			}
		}
	}

	return c.Send(sb.String(), &telebot.SendOptions{ParseMode: telebot.ModeHTML})
}

func handleScan(c telebot.Context, tm *TaskManager, scan usecase.Scan, notifier port.Notifier) error {
	chatID := c.Sender().ID

	if tm.IsBusy(chatID) {
		existing := tm.Get(chatID)
		return c.Send(fmt.Sprintf("⏳ Já tem um job rodando: %s (%v)\nUse /status pra acompanhar.",
			existing.Type, time.Since(existing.StartTime).Round(time.Second)))
	}

	go func() {
		ctx, cancel := context.WithCancel(context.Background())
		tm.Start(chatID, "scan", "Scanner de vagas (scan.mjs + WebSearch L3)", cancel)
		defer tm.End(chatID)

		result, err := scan.Run(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				notifier.Notify(chatID, "🛑 Scan cancelado.")
			} else {
				notifier.Notify(chatID, fmt.Sprintf("❌ Erro no scan: %v", err))
			}
			return
		}

		res := fmt.Sprintf("✅ <b>Scan concluído</b>\n\n%s", result.Summary)
		if result.Pipeline != nil && len(result.Pipeline.Pending) > 0 {
			res += fmt.Sprintf("\n\n<b>Novas no pipeline: %d</b>\n", len(result.Pipeline.Pending))
			for i, item := range result.Pipeline.Pending {
				if i >= 5 {
					res += fmt.Sprintf("\n... +%d mais\n", len(result.Pipeline.Pending)-5)
					break
				}
				res += fmt.Sprintf("  • %s | %s\n", item.Company, item.Title)
			}
		}
		notifier.NotifyHTML(chatID, res)
	}()

	return c.Send("🔎 <b>Scan iniciado.</b> Pode levar alguns minutos.\nUse /status pra acompanhar.",
		&telebot.SendOptions{ParseMode: telebot.ModeHTML})
}

func handleEval(c telebot.Context, tm *TaskManager, evaluate usecase.Evaluate, notifier port.Notifier, url string) error {
	chatID := c.Sender().ID

	if tm.IsBusy(chatID) {
		existing := tm.Get(chatID)
		return c.Send(fmt.Sprintf("⏳ Já tem um job rodando: %s (%v)\nUse /status pra acompanhar.",
			existing.Type, time.Since(existing.StartTime).Round(time.Second)))
	}

	go func() {
		ctx, cancel := context.WithCancel(context.Background())
		tm.Start(chatID, "avaliação", url, cancel)
		defer tm.End(chatID)

		result, err := evaluate.Run(ctx, url)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				notifier.Notify(chatID, "🛑 Avaliação cancelada.")
			} else {
				notifier.Notify(chatID, fmt.Sprintf("❌ Erro na avaliação: %v", err))
			}
			return
		}

		if result.Report != nil {
			latest := result.Report
			res := fmt.Sprintf("✅ <b>Avaliação concluída</b>\n\n"+
				"<b>#%d</b> %s — %s\n"+
				"<b>Score:</b> %s\n"+
				"<b>Veredito:</b> %s",
				latest.Num, latest.Company, latest.Role, latest.Score, latest.Veredito)
			if latest.Archetype != "" {
				res += fmt.Sprintf("\n<b>Arquétipo:</b> %s", latest.Archetype)
			}
			res += fmt.Sprintf("\n\n/report %d pra ver completo", latest.Num)
			notifier.NotifyHTML(chatID, res)
		} else {
			notifier.NotifyHTML(chatID, fmt.Sprintf("✅ <b>Avaliação concluída</b>\n\n%s", result.RawOutput))
		}
	}()

	return c.Send(fmt.Sprintf("📊 <b>Avaliando</b> %s\nUse /status pra acompanhar.", url),
		&telebot.SendOptions{ParseMode: telebot.ModeHTML})
}

func handleCancel(c telebot.Context, tm *TaskManager) error {
	task := tm.Get(c.Sender().ID)
	if task == nil {
		return c.Send("✅ Nenhum job rodando pra cancelar.")
	}

	markup := &telebot.ReplyMarkup{}
	btnYes := markup.Data("🛑 Sim, cancelar", btnCancelYes)
	btnNo := markup.Data("Não", btnCancelNo)
	markup.Inline(markup.Row(btnYes, btnNo))

	return c.Send(fmt.Sprintf("⚠️ Cancelar o job em execução (<b>%s</b>)?", task.Type),
		&telebot.SendOptions{ParseMode: telebot.ModeHTML}, markup)
}

const (
	btnCancelYes = "cancel_yes"
	btnCancelNo  = "cancel_no"
)

func handleCancelConfirm(c telebot.Context, tm *TaskManager) error {
	chatID := c.Sender().ID
	if tm.Cancel(chatID) {
		if err := c.Edit("🛑 Cancelando job..."); err != nil {
			return err
		}
	} else {
		if err := c.Edit("✅ Nenhum job rodando."); err != nil {
			return err
		}
	}
	return c.Respond()
}

func handleCancelAbort(c telebot.Context) error {
	if err := c.Edit("Ok, job continua rodando."); err != nil {
		return err
	}
	return c.Respond()
}

func handleSyncNeeded(c telebot.Context) error {
	return c.Send("⚠️ Dados não encontrados. Rode:\n\n<code>go run ./cmd/bot sync</code>\n\nno diretório do bot para exportar os dados do career-ops.",
		&telebot.SendOptions{ParseMode: telebot.ModeHTML})
}

// ── Ask (NL queries about job data) ──────────────────────────────────────

func handleAsk(c telebot.Context, tm *TaskManager, ask usecase.Ask, notifier port.Notifier, question string) error {
	chatID := c.Sender().ID

	if tm.IsBusy(chatID) {
		existing := tm.Get(chatID)
		return c.Send(fmt.Sprintf("⏳ Já tem um job rodando: %s (%v)\nUse /status pra acompanhar.",
			existing.Type, time.Since(existing.StartTime).Round(time.Second)))
	}

	go func() {
		ctx, cancel := context.WithCancel(context.Background())
		tm.Start(chatID, "consulta", question, cancel)
		defer tm.End(chatID)

		answer, err := ask.Run(ctx, question)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				notifier.Notify(chatID, "🛑 Consulta cancelada.")
			} else {
				notifier.Notify(chatID, fmt.Sprintf("❌ Erro: %v", err))
			}
			return
		}

		notifier.NotifyHTML(chatID, answer)
	}()

	return c.Send(fmt.Sprintf("💬 <b>Consultando:</b> %s", question),
		&telebot.SendOptions{ParseMode: telebot.ModeHTML})
}
