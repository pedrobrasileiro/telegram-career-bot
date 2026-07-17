package main

import (
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"gopkg.in/telebot.v3"
)

func SetupBot(cfg *Config, tm *TaskManager) (*telebot.Bot, error) {
	pref := telebot.Settings{
		Token:  cfg.BotToken,
		Poller: &telebot.LongPoller{Timeout: 10 * time.Second},
	}

	bot, err := telebot.NewBot(pref)
	if err != nil {
		return nil, fmt.Errorf("criando bot: %w", err)
	}

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
		return handleScan(c, cfg, tm)
	})

	bot.Handle("/status", func(c telebot.Context) error {
		return handleStatus(c, tm)
	})

	bot.Handle("/vstatus", func(c telebot.Context) error {
		return handleStatus(c, tm)
	})

	ai := NewAwaitingInput()

	bot.Handle("/tracker", func(c telebot.Context) error {
		filter := strings.TrimSpace(strings.TrimPrefix(c.Text(), "/tracker"))
		if filter == "" {
			ai.Set(c.Sender().ID, "tracker")
			return c.Send("Digite o status pra filtrar (ex: Interview) ou \"todos\":", &telebot.ReplyMarkup{ForceReply: true})
		}
		return handleTrackerFiltered(c, cfg, filter)
	})

	bot.Handle("/stats", func(c telebot.Context) error {
		return handleStats(c, cfg)
	})

	bot.Handle("/pipeline", func(c telebot.Context) error {
		return handlePipeline(c, cfg)
	})

	bot.Handle("/followup", func(c telebot.Context) error {
		return handleFollowUp(c, cfg)
	})

	bot.Handle("/report", func(c telebot.Context) error {
		args := strings.Fields(c.Text())
		if len(args) < 2 {
			ai.Set(c.Sender().ID, "report")
			return c.Send("Digite o número do report (ex: 194):", &telebot.ReplyMarkup{ForceReply: true})
		}
		return handleReport(c, cfg)
	})

	bot.Handle("/agenda", func(c telebot.Context) error {
		return handleAgenda(c, cfg)
	})

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
				return handleReportByNum(c, cfg, num)
			case "tracker":
				filter := strings.TrimSpace(text)
				if strings.EqualFold(filter, "todos") {
					filter = ""
				}
				return handleTrackerFiltered(c, cfg, filter)
			}
		}

		if isURL(text) {
			return handleEval(c, cfg, tm, text)
		}
		if strings.HasPrefix(text, "/") {
			return nil // unknown command, ignore
		}
		return handleAsk(c, cfg, tm, text)
	})

	if err := bot.SetCommands([]telebot.Command{
		{Text: "scan", Description: "Scan completo de vagas"},
		{Text: "status", Description: "Job ativo em execução"},
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
<b>/tracker [status]</b> — Resumo do tracker
<b>/stats</b> — Funil completo
<b>/pipeline</b> — URLs pendentes
<b>/followup</b> — Follow-ups próximos 7 dias
<b>/report 194</b> — Resumo de um report
<b>/agenda</b> — Próximas entrevistas`
}

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

func handleTracker(c telebot.Context, cfg *Config) error {
	filter := strings.TrimSpace(strings.TrimPrefix(c.Text(), "/tracker"))
	return handleTrackerFiltered(c, cfg, filter)
}

func handleTrackerFiltered(c telebot.Context, cfg *Config, filter string) error {
	trackerData := loadTrackerData(cfg.DataPath)
	if trackerData == nil {
		return handleSyncNeeded(c)
	}

	statusCounts := make(map[string]int)
	for _, a := range trackerData.Applications {
		statusCounts[a.Status]++
	}

	var filtered []Application
	if filter != "" {
		for _, a := range trackerData.Applications {
			if strings.EqualFold(a.Status, filter) {
				filtered = append(filtered, a)
			}
		}
	} else {
		filtered = trackerData.Applications
	}

	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].Num > filtered[j].Num
	})

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("<b>📊 Tracker</b> (%d vagas)\n\n", len(trackerData.Applications)))
	sb.WriteString("<b>Por status:</b>\n")
	for _, s := range statusOrder() {
		if count, ok := statusCounts[s]; ok && count > 0 {
			sb.WriteString(fmt.Sprintf("  %s %s: %d\n", emojiForTrackerStatus(s), s, count))
		}
	}

	if filter != "" {
		sb.WriteString(fmt.Sprintf("\n<b>Filtro: %s</b>\n", filter))
	}

	limit := 5
	if len(filtered) < limit {
		limit = len(filtered)
	}
	sb.WriteString(fmt.Sprintf("\n<b>Últimas %d:</b>\n", limit))
	for i := 0; i < limit; i++ {
		a := filtered[i]
		emoji := emojiForTrackerStatus(a.Status)
		sb.WriteString(fmt.Sprintf("%s <b>#%d</b> %s — %s (%s)\n", emoji, a.Num, a.Company, a.Role, a.Status))
	}

	if len(filtered) > limit {
		sb.WriteString(fmt.Sprintf("\n... +%d mais\n", len(filtered)-limit))
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

func statusOrder() []string {
	return []string{"Interview", "Applied", "Offer", "Evaluated", "Responded", "Rejected", "Discarded", "SKIP"}
}

func handleStats(c telebot.Context, cfg *Config) error {
	stats := loadStatsData(cfg.DataPath)
	if stats == nil {
		return handleSyncNeeded(c)
	}

	tracker := loadTrackerData(cfg.DataPath)
	if tracker == nil {
		return handleSyncNeeded(c)
	}

	byStatus := stats.ByStatus

	everApplied := byStatus["Applied"] + byStatus["Interview"] + byStatus["Offer"] + byStatus["Rejected"] + byStatus["Responded"]
	everResponded := byStatus["Interview"] + byStatus["Offer"] + byStatus["Rejected"] + byStatus["Responded"]
	interview := byStatus["Interview"] + byStatus["Offer"]
	offer := byStatus["Offer"]
	rejected := byStatus["Rejected"]
	active := byStatus["Applied"] + byStatus["Interview"] + byStatus["Responded"] + byStatus["Evaluated"]
	skipped := byStatus["Discarded"] + byStatus["SKIP"]

	respRate := 0.0
	if everApplied > 0 {
		respRate = float64(everResponded) / float64(everApplied) * 100
	}
	interviewRate := 0.0
	if everApplied > 0 {
		interviewRate = float64(interview) / float64(everApplied) * 100
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("<b>📈 Stats</b>\n\n"))
	sb.WriteString(fmt.Sprintf("<b>Total:</b> %d | <b>Ativas:</b> %d\n", stats.Total, active))
	sb.WriteString(fmt.Sprintf("<b>Avg fit:</b> %.1f/5\n\n", stats.AverageScore))

	sb.WriteString("<b>Por status:</b>\n")
	for _, s := range statusOrder() {
		if c, ok := byStatus[s]; ok && c > 0 {
			sb.WriteString(fmt.Sprintf("  %s %s: %d\n", emojiForTrackerStatus(s), s, c))
		}
	}

	sb.WriteString(fmt.Sprintf("\n<b>Funil:</b>\n"))
	sb.WriteString(fmt.Sprintf("  📤 Aplicadas: %d\n", everApplied))
	sb.WriteString(fmt.Sprintf("  📩 Responderam: %d (%.0f%%)\n", everResponded, respRate))
	sb.WriteString(fmt.Sprintf("  🎯 Entrevista: %d (%.0f%%)\n", interview, interviewRate))
	sb.WriteString(fmt.Sprintf("  🎉 Oferta: %d\n", offer))
	sb.WriteString(fmt.Sprintf("  ❌ Rejeitadas: %d\n", rejected))
	sb.WriteString(fmt.Sprintf("  🗑 Descartadas/Skip: %d\n", skipped))

	return c.Send(sb.String(), &telebot.SendOptions{ParseMode: telebot.ModeHTML})
}

func handlePipeline(c telebot.Context, cfg *Config) error {
	pipeline := loadPipelineData(cfg.DataPath)
	if pipeline == nil {
		return handleSyncNeeded(c)
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("<b>📥 Pipeline</b>\n\n"))

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

func handleFollowUp(c telebot.Context, cfg *Config) error {
	followUpData := loadFollowUpData(cfg.DataPath)
	if followUpData == nil {
		return handleSyncNeeded(c)
	}

	now := time.Now()
	nextWeek := now.Add(7 * 24 * time.Hour)

	var upcoming []FollowUpItem
	for _, item := range followUpData.Items {
		d, err := time.Parse("2006-01-02", item.Date)
		if err != nil {
			continue
		}
		if d.After(now) && d.Before(nextWeek) {
			upcoming = append(upcoming, item)
		}
	}

	sort.Slice(upcoming, func(i, j int) bool {
		return upcoming[i].Date < upcoming[j].Date
	})

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("<b>📅 Follow-ups — próximos 7 dias</b>\n\n"))

	if len(upcoming) == 0 {
		sb.WriteString("Nenhum follow-up nos próximos 7 dias.")
	} else {
		for _, item := range upcoming {
			sb.WriteString(fmt.Sprintf("<b>%s</b> — %s\n", item.Date, item.Company))
			if item.Action != "" {
				sb.WriteString(fmt.Sprintf("  Ação: %s\n", item.Action))
			}
		}
	}

	return c.Send(sb.String(), &telebot.SendOptions{ParseMode: telebot.ModeHTML})
}

func handleReport(c telebot.Context, cfg *Config) error {
	args := strings.Fields(c.Text())
	if len(args) < 2 {
		return c.Send("Uso: /report <número>\nEx: /report 194")
	}

	num, err := parseInt(args[1])
	if err != nil {
		return c.Send("Número inválido.")
	}

	return handleReportByNum(c, cfg, num)
}

func handleReportByNum(c telebot.Context, cfg *Config, num int) error {
	reportsData := loadReportsIndexData(cfg.DataPath)
	if reportsData == nil {
		return handleSyncNeeded(c)
	}

	var report *ReportSummary
	for _, r := range reportsData.Reports {
		if r.Num == num {
			report = &r
			break
		}
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

	reportsDir := filepathFromCareerOps(cfg.CareerOpsPath, "reports")
	sb.WriteString(fmt.Sprintf("\n<b>Arquivo:</b> %s/%s", reportsDir, report.Filename))

	return c.Send(sb.String(), &telebot.SendOptions{ParseMode: telebot.ModeHTML})
}

func handleAgenda(c telebot.Context, cfg *Config) error {
	trackerData := loadTrackerData(cfg.DataPath)
	if trackerData == nil {
		return handleSyncNeeded(c)
	}

	var interviews []Application
	for _, a := range trackerData.Applications {
		if a.Status == "Interview" {
			interviews = append(interviews, a)
		}
	}

	sort.Slice(interviews, func(i, j int) bool {
		return interviews[i].Num > interviews[j].Num
	})

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("<b>🎯 Agenda de Entrevistas</b>\n\n"))

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

func handleScan(c telebot.Context, cfg *Config, tm *TaskManager) error {
	chatID := c.Sender().ID

	if tm.IsBusy(chatID) {
		existing := tm.Get(chatID)
		return c.Send(fmt.Sprintf("⏳ Já tem um job rodando: %s (%v)\nUse /status pra acompanhar.",
			existing.Type, time.Since(existing.StartTime).Round(time.Second)))
	}

	go func() {
		tm.Start(chatID, "scan", "Scanner de vagas (scan.mjs + WebSearch L3)")
		defer tm.End(chatID)

		out, err := runOpenCode(cfg.CareerOpsPath, "Run career-ops scan mode", cfg.ScanTimeout)
		if err != nil {
			bot := c.Bot()
			bot.Send(c.Sender(), fmt.Sprintf("❌ Erro no scan: %v", err))
			return
		}

		exportAll(cfg.CareerOpsPath, cfg.DataPath)

		summary := extractScanSummary(out)
		pipeline := loadPipelineData(cfg.DataPath)

		res := fmt.Sprintf("✅ <b>Scan concluído</b>\n\n%s", summary)
		if pipeline != nil && len(pipeline.Pending) > 0 {
			res += fmt.Sprintf("\n\n<b>Novas no pipeline: %d</b>\n", len(pipeline.Pending))
			for i, item := range pipeline.Pending {
				if i >= 5 {
					res += fmt.Sprintf("\n... +%d mais\n", len(pipeline.Pending)-5)
					break
				}
				res += fmt.Sprintf("  • %s | %s\n", item.Company, item.Title)
			}
		}
		bot := c.Bot()
		bot.Send(c.Sender(), res, &telebot.SendOptions{ParseMode: telebot.ModeHTML})
	}()

	return c.Send("🔎 <b>Scan iniciado.</b> Pode levar alguns minutos.\nUse /status pra acompanhar.",
		&telebot.SendOptions{ParseMode: telebot.ModeHTML})
}

func handleEval(c telebot.Context, cfg *Config, tm *TaskManager, url string) error {
	chatID := c.Sender().ID

	if tm.IsBusy(chatID) {
		existing := tm.Get(chatID)
		return c.Send(fmt.Sprintf("⏳ Já tem um job rodando: %s (%v)\nUse /status pra acompanhar.",
			existing.Type, time.Since(existing.StartTime).Round(time.Second)))
	}

	go func() {
		tm.Start(chatID, "avaliação", url)
		defer tm.End(chatID)

		prompt := fmt.Sprintf("Evaluate this JD with career-ops auto-pipeline: %s", url)
		out, err := runOpenCode(cfg.CareerOpsPath, prompt, cfg.EvalTimeout)
		if err != nil {
			bot := c.Bot()
			bot.Send(c.Sender(), fmt.Sprintf("❌ Erro na avaliação: %v", err))
			return
		}

		exportAll(cfg.CareerOpsPath, cfg.DataPath)

		reportsData := loadReportsIndexData(cfg.DataPath)
		bot := c.Bot()

		if reportsData != nil && len(reportsData.Reports) > 0 {
			latest := reportsData.Reports[0]
			res := fmt.Sprintf("✅ <b>Avaliação concluída</b>\n\n"+
				"<b>#%d</b> %s — %s\n"+
				"<b>Score:</b> %s\n"+
				"<b>Veredito:</b> %s",
				latest.Num, latest.Company, latest.Role, latest.Score, latest.Veredito)
			if latest.Archetype != "" {
				res += fmt.Sprintf("\n<b>Arquétipo:</b> %s", latest.Archetype)
			}
			res += fmt.Sprintf("\n\n/report %d pra ver completo", latest.Num)
			bot.Send(c.Sender(), res, &telebot.SendOptions{ParseMode: telebot.ModeHTML})
		} else {
			lastLine := lastNonEmptyLine(out)
			bot.Send(c.Sender(), fmt.Sprintf("✅ <b>Avaliação concluída</b>\n\n%s", lastLine),
				&telebot.SendOptions{ParseMode: telebot.ModeHTML})
		}
	}()

	return c.Send(fmt.Sprintf("📊 <b>Avaliando</b> %s\nUse /status pra acompanhar.", url),
		&telebot.SendOptions{ParseMode: telebot.ModeHTML})
}

func extractScanSummary(out string) string {
	lines := strings.Split(out, "\n")
	var summaryLines []string
	inSummary := false
	for _, line := range lines {
		if strings.Contains(line, "Portal Scan") || strings.Contains(line, "Offer") || strings.Contains(line, "New added") {
			inSummary = true
		}
		if inSummary {
			summaryLines = append(summaryLines, line)
		}
	}
	if len(summaryLines) == 0 {
		return "Scan concluído. Confira o pipeline."
	}
	return strings.Join(summaryLines, "\n")
}

func handleSyncNeeded(c telebot.Context) error {
	return c.Send("⚠️ Dados não encontrados. Rode:\n\n<code>go run . sync</code>\n\nno diretório do bot para exportar os dados do career-ops.",
		&telebot.SendOptions{ParseMode: telebot.ModeHTML})
}

// ── Ask (NL queries about job data) ──────────────────────────────────────

func handleAsk(c telebot.Context, cfg *Config, tm *TaskManager, question string) error {
	chatID := c.Sender().ID

	if tm.IsBusy(chatID) {
		existing := tm.Get(chatID)
		return c.Send(fmt.Sprintf("⏳ Já tem um job rodando: %s (%v)\nUse /status pra acompanhar.",
			existing.Type, time.Since(existing.StartTime).Round(time.Second)))
	}

	go func() {
		tm.Start(chatID, "consulta", question)
		defer tm.End(chatID)

		ctx := buildDataContext(cfg.DataPath)
		if ctx == "" {
			ctx = "(dados ainda não exportados — rode go run . sync)"
		}

		prompt := fmt.Sprintf("Você é um assistente de job search. Responda a pergunta do usuário em português brasileiro (pt-BR) de forma concisa e direta, usando apenas os dados fornecidos abaixo. Se os dados não forem suficientes para responder, diga exatamente o que falta.\n\n## DADOS DO TRACKER\n\n%s\n\n## PERGUNTA DO USUÁRIO\n\n%s", ctx, question)

		out, err := runOpenCode(cfg.CareerOpsPath, prompt, 120*time.Second)
		if err != nil {
			bot := c.Bot()
			bot.Send(c.Sender(), fmt.Sprintf("❌ Erro: %v", err))
			return
		}

		answer := strings.TrimSpace(out)
		if len(answer) > 4000 {
			answer = answer[:4000] + "\n\n…"
		}

		bot := c.Bot()
		bot.Send(c.Sender(), answer, &telebot.SendOptions{ParseMode: telebot.ModeHTML})
	}()

	return c.Send(fmt.Sprintf("💬 <b>Consultando:</b> %s", question),
		&telebot.SendOptions{ParseMode: telebot.ModeHTML})
}

func buildDataContext(dataPath string) string {
	tracker := loadTrackerData(dataPath)
	stats := loadStatsData(dataPath)
	pipeline := loadPipelineData(dataPath)

	if tracker == nil || stats == nil {
		return ""
	}

	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Total de vagas: %d\n", stats.Total))
	sb.WriteString(fmt.Sprintf("Score médio: %.1f/5\n\n", stats.AverageScore))

	sb.WriteString("Por status:\n")
	for _, s := range statusOrder() {
		if c, ok := stats.ByStatus[s]; ok && c > 0 {
			sb.WriteString(fmt.Sprintf("  %s: %d\n", s, c))
		}
	}

	if pipeline != nil {
		sb.WriteString(fmt.Sprintf("\nPipeline: %d pendentes\n", len(pipeline.Pending)))
	}

	sb.WriteString("\nÚltimas 20 vagas (ordenadas por #):\n")
	sorted := make([]Application, len(tracker.Applications))
	copy(sorted, tracker.Applications)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Num > sorted[j].Num
	})
	limit := 20
	if len(sorted) < limit {
		limit = len(sorted)
	}
	for i := 0; i < limit; i++ {
		a := sorted[i]
		sb.WriteString(fmt.Sprintf("#%d %s | %s | %s | %s", a.Num, a.Date, a.Company, a.Role, a.Status))
		if a.Score != "" && a.Score != "N/A" {
			sb.WriteString(fmt.Sprintf(" | %s", a.Score))
		}
		if a.Notes != "" {
			notes := a.Notes
			if len(notes) > 120 {
				notes = notes[:120] + "..."
			}
			sb.WriteString(fmt.Sprintf(" | %s", notes))
		}
		sb.WriteByte('\n')
	}

	return sb.String()
}
