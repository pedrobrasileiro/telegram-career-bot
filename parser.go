package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

type Application struct {
	Num        int    `json:"num"`
	Date       string `json:"date"`
	Company    string `json:"company"`
	Role       string `json:"role"`
	Score      string `json:"score"`
	Status     string `json:"status"`
	PDF        string `json:"pdf"`
	ReportPath string `json:"reportPath"`
	Notes      string `json:"notes"`
	Via        string `json:"via,omitempty"`
	Location   string `json:"location,omitempty"`
}

type Pipeline struct {
	Pending   []PipelineItem `json:"pending"`
	Processed []PipelineItem `json:"processed"`
}

type PipelineItem struct {
	URL     string `json:"url"`
	Company string `json:"company"`
	Title   string `json:"title"`
}

type FollowUpItem struct {
	Date    string `json:"date"`
	Company string `json:"company"`
	Role    string `json:"role"`
	Action  string `json:"action"`
	Notes   string `json:"notes"`
}

type ReportSummary struct {
	Num        int    `json:"num"`
	Title      string `json:"title"`
	Company    string `json:"company"`
	Role       string `json:"role"`
	Filename   string `json:"filename"`
	Date       string `json:"date"`
	Score      string `json:"score"`
	Veredito   string `json:"veredito"`
	URL        string `json:"url"`
	Legitimacy string `json:"legitimacy"`
	Archetype  string `json:"archetype"`
}

type Stats struct {
	Total        int            `json:"total"`
	ByStatus     map[string]int `json:"byStatus"`
	AverageScore float64        `json:"averageScore"`
	Funnel       Funnel         `json:"funnel"`
	ExportedAt   string         `json:"exportedAt"`
}

type Funnel struct {
	Total     int `json:"total"`
	Evaluated int `json:"evaluated"`
	Applied   int `json:"applied"`
	Interview int `json:"interview"`
	Offer     int `json:"offer"`
	Rejected  int `json:"rejected"`
	Discarded int `json:"discarded"`
	Skipped   int `json:"skipped"`
}

type TrackerData struct {
	Applications []Application `json:"applications"`
	ExportedAt   string        `json:"exportedAt"`
}

type FollowUpData struct {
	Items      []FollowUpItem `json:"items"`
	ExportedAt string         `json:"exportedAt"`
}

type ReportsIndexData struct {
	Reports    []ReportSummary `json:"reports"`
	ExportedAt string          `json:"exportedAt"`
}

func parseApplications(careerOpsPath string) ([]Application, error) {
	f, err := os.Open(filepath.Join(careerOpsPath, "data", "applications.md"))
	if err != nil {
		return nil, fmt.Errorf("erro abrindo applications.md: %w", err)
	}
	defer f.Close()

	var apps []Application
	scanner := bufio.NewScanner(f)
	inTable := false
	colCount := 0

	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "| # |") {
			inTable = true
			continue
		}
		if inTable && strings.Contains(line, "|---") {
			colCount = strings.Count(line, "|") - 1
			continue
		}
		if !inTable || !strings.HasPrefix(line, "|") {
			continue
		}

		cols := parseTableRow(line, colCount)
		if len(cols) < 8 {
			continue
		}

		num, _ := strconv.Atoi(strings.TrimSpace(cols[0]))

		app := Application{
			Num:     num,
			Date:    strings.TrimSpace(cols[1]),
			Company: strings.TrimSpace(cols[2]),
			Role:    strings.TrimSpace(cols[3]),
			Score:   strings.TrimSpace(cols[4]),
			Status:  cleanStatus(strings.TrimSpace(cols[5])),
			PDF:     strings.TrimSpace(cols[6]),
			Notes:   "",
		}

		reportCell := strings.TrimSpace(cols[7])
		app.ReportPath = extractReportPath(reportCell)

		if len(cols) > 8 {
			noteContent := strings.TrimSpace(cols[8])
			app.Notes, app.Via, app.Location = parseNotes(noteContent)
		}

		apps = append(apps, app)
	}

	return apps, scanner.Err()
}

func parseTableRow(line string, expectedCols int) []string {
	parts := strings.Split(line, "|")
	cells := make([]string, 0, expectedCols)
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" && len(cells) == 0 {
			continue
		}
		cells = append(cells, p)
	}
	if len(cells) > 0 && cells[len(cells)-1] == "" {
		cells = cells[:len(cells)-1]
	}
	if len(cells) > expectedCols && expectedCols > 0 {
		last := strings.Join(cells[expectedCols-1:], " | ")
		cells = append(cells[:expectedCols-1], last)
	}
	return cells
}

func cleanStatus(s string) string {
	s = strings.TrimPrefix(s, "**")
	s = strings.TrimSuffix(s, "**")
	s = strings.TrimSpace(s)
	return s
}

func extractReportPath(cell string) string {
	re := regexp.MustCompile(`\(\.\.?/?(reports/[^)]+)\)`)
	matches := re.FindStringSubmatch(cell)
	if len(matches) >= 2 {
		return matches[1]
	}
	return ""
}

func parseNotes(raw string) (notes, via, location string) {
	notes = raw

	viaRe := regexp.MustCompile(`via=(\S+)`)
	if m := viaRe.FindStringSubmatch(raw); len(m) >= 2 {
		via = m[1]
	}

	return notes, via, location
}

func parsePipeline(careerOpsPath string) (*Pipeline, error) {
	f, err := os.Open(filepath.Join(careerOpsPath, "data", "pipeline.md"))
	if err != nil {
		return nil, err
	}
	defer f.Close()

	p := &Pipeline{}
	re := regexp.MustCompile(`^[-*]\s*\[[ x]\]\s*(.+?)\s*\|\s*(.+?)\s*\|\s*(.+)$`)

	scanner := bufio.NewScanner(f)
	inPending := false
	inProcessed := false

	for scanner.Scan() {
		line := scanner.Text()

		if strings.Contains(strings.ToLower(line), "pendente") || strings.Contains(strings.ToLower(line), "pending") {
			inPending = true
			inProcessed = false
			continue
		}
		if strings.Contains(strings.ToLower(line), "processado") || strings.Contains(strings.ToLower(line), "processed") {
			inProcessed = true
			inPending = false
			continue
		}

		matches := re.FindStringSubmatch(line)
		if len(matches) >= 4 {
			item := PipelineItem{
				URL:     strings.TrimSpace(matches[1]),
				Company: strings.TrimSpace(matches[2]),
				Title:   strings.TrimSpace(matches[3]),
			}
			if inProcessed {
				p.Processed = append(p.Processed, item)
			} else if inPending {
				p.Pending = append(p.Pending, item)
			} else {
				p.Pending = append(p.Pending, item)
			}
		}
	}

	return p, scanner.Err()
}

func parseFollowUps(careerOpsPath string) ([]FollowUpItem, error) {
	f, err := os.Open(filepath.Join(careerOpsPath, "data", "follow-ups.md"))
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var items []FollowUpItem
	scanner := bufio.NewScanner(f)
	inTable := false
	colCount := 0

	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "| # |") && strings.Contains(line, "App#") {
			inTable = true
			continue
		}
		if inTable && strings.Contains(line, "|---") {
			colCount = strings.Count(line, "|") - 1
			continue
		}
		if !inTable || !strings.HasPrefix(line, "|") {
			continue
		}

		cols := parseTableRow(line, colCount)
		if len(cols) < 4 {
			continue
		}

		item := FollowUpItem{
			Date:    "",
			Company: "",
		}
		// Columns: # | App# | Date | Company | Role | Channel | Contact | Notes
		if len(cols) > 2 {
			item.Date = cleanFollowUpDate(strings.TrimSpace(cols[2]))
		}
		if len(cols) > 3 {
			item.Company = strings.TrimSpace(cols[3])
		}
		if len(cols) > 4 {
			item.Role = strings.TrimSpace(cols[4])
		}
		if len(cols) > 5 {
			item.Action = strings.TrimSpace(cols[5])
		}
		if len(cols) > 7 {
			item.Notes = strings.TrimSpace(cols[7])
		} else if len(cols) > 6 {
			item.Notes = strings.TrimSpace(cols[6])
		}

		items = append(items, item)
	}

	return items, scanner.Err()
}

func cleanFollowUpDate(date string) string {
	date = strings.TrimPrefix(date, "**")
	date = strings.TrimSuffix(date, "**")
	idx := strings.Index(date, "(")
	if idx > 0 {
		date = strings.TrimSpace(date[:idx])
	}
	return date
}

func parseReportHeaders(careerOpsPath string) ([]ReportSummary, error) {
	reportsDir := filepath.Join(careerOpsPath, "reports")
	entries, err := os.ReadDir(reportsDir)
	if err != nil {
		return nil, err
	}

	titleNumRe := regexp.MustCompile(`^#\s*(\d+)\s*[—–-]\s*(.+?)\s*\|\s*(.+)$`)
	titleEvalRe := regexp.MustCompile(`^#\s*Evaluation:\s*(.+?)\s*[—–-]\s*(.+)$`)
	scoreRe := regexp.MustCompile(`\*\*Score:\*\*\s*(.+)`)
	vereditoHeaderRe := regexp.MustCompile(`^##\s*Veredito\s*$`)
	vereditoInlineRe := regexp.MustCompile(`\*\*Veredito:\*\*\s*(.+)`)
	urlRe := regexp.MustCompile(`\*\*URL:\*\*\s*(.+)`)
	legitimacyRe := regexp.MustCompile(`\*\*Legitimacy:\*\*\s*(.+)`)
	archetypeRe := regexp.MustCompile(`\*\*Archetype:\*\*\s*(.+)`)
	filenameNumRe := regexp.MustCompile(`^(\d+)-`)
	filenameDateRe := regexp.MustCompile(`(\d{4}-\d{2}-\d{2})\.md$`)

	var reports []ReportSummary

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		f, err := os.Open(filepath.Join(reportsDir, entry.Name()))
		if err != nil {
			continue
		}

		var r ReportSummary
		r.Filename = entry.Name()

		if m := filenameNumRe.FindStringSubmatch(entry.Name()); len(m) >= 2 {
			r.Num, _ = strconv.Atoi(m[1])
		}

		if m := filenameDateRe.FindStringSubmatch(entry.Name()); len(m) >= 2 {
			r.Date = m[1]
		}

		inVeredito := false
		scanner := bufio.NewScanner(f)
		for scanner.Scan() && (r.Score == "" || r.Veredito == "" || r.URL == "" || r.Title == "") {
			line := scanner.Text()

			if m := titleNumRe.FindStringSubmatch(line); len(m) >= 4 {
				if r.Num == 0 {
					r.Num, _ = strconv.Atoi(m[1])
				}
				r.Company = strings.TrimSpace(m[2])
				r.Role = strings.TrimSpace(m[3])
				r.Title = fmt.Sprintf("%s | %s", r.Company, r.Role)
				continue
			}

			if m := titleEvalRe.FindStringSubmatch(line); len(m) >= 3 && r.Title == "" {
				r.Company = strings.TrimSpace(m[1])
				r.Role = strings.TrimSpace(m[2])
				r.Title = fmt.Sprintf("%s | %s", r.Company, r.Role)
				continue
			}

			if m := scoreRe.FindStringSubmatch(line); len(m) >= 2 {
				r.Score = strings.TrimSpace(m[1])
				continue
			}

			if m := urlRe.FindStringSubmatch(line); len(m) >= 2 {
				r.URL = strings.TrimSpace(m[1])
				continue
			}

			if m := legitimacyRe.FindStringSubmatch(line); len(m) >= 2 {
				r.Legitimacy = strings.TrimSpace(m[1])
				continue
			}

			if m := archetypeRe.FindStringSubmatch(line); len(m) >= 2 {
				r.Archetype = strings.TrimSpace(m[1])
				continue
			}

			if m := vereditoInlineRe.FindStringSubmatch(line); len(m) >= 2 && r.Veredito == "" {
				r.Veredito = strings.TrimSpace(m[1])
				continue
			}

			if vereditoHeaderRe.MatchString(line) {
				inVeredito = true
				continue
			}

			if inVeredito && line != "" && !strings.HasPrefix(line, "#") {
				r.Veredito = strings.TrimSpace(line)
				inVeredito = false
			}
		}

		f.Close()

		if r.Num > 0 {
			reports = append(reports, r)
		}
	}

	sort.Slice(reports, func(i, j int) bool {
		return reports[i].Num > reports[j].Num
	})

	return reports, nil
}

func computeStats(apps []Application) Stats {
	now := time.Now().Format("2006-01-02")
	s := Stats{
		ExportedAt: now,
		ByStatus:   make(map[string]int),
	}

	var totalScore float64
	var scoreCount int

	for _, a := range apps {
		s.Total++
		s.ByStatus[a.Status]++

		scoreStr := strings.TrimSuffix(a.Score, "/5")
		if score, err := strconv.ParseFloat(scoreStr, 64); err == nil {
			totalScore += score
			scoreCount++
		}
	}

	if scoreCount > 0 {
		s.AverageScore = totalScore / float64(scoreCount)
	}

	s.Funnel = Funnel{
		Total:     s.Total,
		Evaluated: s.ByStatus["Evaluated"],
		Applied:   s.ByStatus["Applied"],
		Interview: s.ByStatus["Interview"],
		Offer:     s.ByStatus["Offer"],
		Rejected:  s.ByStatus["Rejected"],
		Discarded: s.ByStatus["Discarded"],
		Skipped:   s.ByStatus["SKIP"],
	}

	return s
}

func writeJSON(path string, v interface{}) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
