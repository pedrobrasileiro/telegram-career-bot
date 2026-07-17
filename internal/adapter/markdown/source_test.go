package markdown

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseTableRow(t *testing.T) {
	cols := parseTableRow("| 1 | 2024-01-01 | Acme | Dev | 4/5 | Applied | pdf | (../reports/1-acme.md) |", 8)
	want := []string{"1", "2024-01-01", "Acme", "Dev", "4/5", "Applied", "pdf", "(../reports/1-acme.md)"}
	if len(cols) != len(want) {
		t.Fatalf("got %d cols, want %d: %v", len(cols), len(want), cols)
	}
	for i, w := range want {
		if cols[i] != w {
			t.Errorf("col %d = %q, want %q", i, cols[i], w)
		}
	}
}

func TestParseTableRowExtraColsJoinsTail(t *testing.T) {
	cols := parseTableRow("| 1 | 2 | 3 | extra | more |", 3)
	if len(cols) != 3 {
		t.Fatalf("got %d cols, want 3: %v", len(cols), cols)
	}
	if cols[2] != "3 | extra | more" {
		t.Errorf("last col = %q, want joined tail", cols[2])
	}
}

func TestCleanStatus(t *testing.T) {
	cases := map[string]string{
		"**Interview**": "Interview",
		"Applied":       "Applied",
		"**Offer**":     "Offer",
	}
	for in, want := range cases {
		if got := cleanStatus(in); got != want {
			t.Errorf("cleanStatus(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestExtractReportPath(t *testing.T) {
	cases := map[string]string{
		"[link](../reports/194-acme.md)": "reports/194-acme.md",
		"[link](./reports/194-acme.md)":  "reports/194-acme.md",
		"sem link nenhum":                "",
	}
	for in, want := range cases {
		if got := extractReportPath(in); got != want {
			t.Errorf("extractReportPath(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestParseNotes(t *testing.T) {
	notes, via, location := parseNotes("Recrutador ativo via=linkedin")
	if notes != "Recrutador ativo via=linkedin" {
		t.Errorf("notes = %q", notes)
	}
	if via != "linkedin" {
		t.Errorf("via = %q, want linkedin", via)
	}
	if location != "" {
		t.Errorf("location = %q, want empty (não implementado)", location)
	}
}

func TestCleanFollowUpDate(t *testing.T) {
	cases := map[string]string{
		"**2024-01-01**":        "2024-01-01",
		"2024-01-01 (atrasado)": "2024-01-01",
		"2024-01-01":            "2024-01-01",
	}
	for in, want := range cases {
		if got := cleanFollowUpDate(in); got != want {
			t.Errorf("cleanFollowUpDate(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestParseApplications(t *testing.T) {
	dir := t.TempDir()
	dataDir := filepath.Join(dir, "data")
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		t.Fatal(err)
	}
	content := `# Applications

| # | Date | Company | Role | Score | Status | PDF | Report | Notes |
|---|------|---------|------|-------|--------|-----|--------|-------|
| 1 | 2024-01-01 | Acme | Dev | 4/5 | **Applied** | cv.pdf | [link](../reports/1-acme.md) | via=linkedin |
| 2 | 2024-01-02 | Globex | SRE | 3/5 | **Interview** | cv.pdf | [link](../reports/2-globex.md) | |
`
	if err := os.WriteFile(filepath.Join(dataDir, "applications.md"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	apps, err := New(dir).ParseApplications()
	if err != nil {
		t.Fatalf("ParseApplications erro: %v", err)
	}
	if len(apps) != 2 {
		t.Fatalf("got %d apps, want 2: %+v", len(apps), apps)
	}
	if apps[0].Num != 1 || apps[0].Company != "Acme" || apps[0].Status != "Applied" {
		t.Errorf("app[0] = %+v", apps[0])
	}
	if apps[0].Via != "linkedin" {
		t.Errorf("app[0].Via = %q, want linkedin", apps[0].Via)
	}
	if apps[0].ReportPath != "reports/1-acme.md" {
		t.Errorf("app[0].ReportPath = %q", apps[0].ReportPath)
	}
	if apps[1].Status != "Interview" {
		t.Errorf("app[1].Status = %q, want Interview", apps[1].Status)
	}
}

func TestParsePipeline(t *testing.T) {
	dir := t.TempDir()
	dataDir := filepath.Join(dir, "data")
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		t.Fatal(err)
	}
	content := `# Pipeline

## Pendentes
- [ ] https://example.com/vaga1 | Acme | Dev
- [ ] https://example.com/vaga2 | Globex | SRE

## Processados
- [x] https://example.com/vaga3 | Initech | PM
`
	if err := os.WriteFile(filepath.Join(dataDir, "pipeline.md"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	p, err := New(dir).ParsePipeline()
	if err != nil {
		t.Fatalf("ParsePipeline erro: %v", err)
	}
	if len(p.Pending) != 2 {
		t.Fatalf("got %d pending, want 2: %+v", len(p.Pending), p.Pending)
	}
	if len(p.Processed) != 1 {
		t.Fatalf("got %d processed, want 1: %+v", len(p.Processed), p.Processed)
	}
	if p.Pending[0].Company != "Acme" || p.Pending[0].Title != "Dev" {
		t.Errorf("pending[0] = %+v", p.Pending[0])
	}
	if p.Processed[0].Company != "Initech" {
		t.Errorf("processed[0] = %+v", p.Processed[0])
	}
}

func TestParseFollowUps(t *testing.T) {
	dir := t.TempDir()
	dataDir := filepath.Join(dir, "data")
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		t.Fatal(err)
	}
	content := `# Follow-ups

| # | App# | Date | Company | Role | Channel | Contact | Notes |
|---|------|------|---------|------|---------|---------|-------|
| 1 | 5 | **2024-01-10** | Acme | Dev | Email | recrutador | lembrar de mencionar disponibilidade |
`
	if err := os.WriteFile(filepath.Join(dataDir, "follow-ups.md"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	items, err := New(dir).ParseFollowUps()
	if err != nil {
		t.Fatalf("ParseFollowUps erro: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("got %d items, want 1: %+v", len(items), items)
	}
	if items[0].Date != "2024-01-10" || items[0].Company != "Acme" || items[0].Role != "Dev" {
		t.Errorf("item[0] = %+v", items[0])
	}
}

func TestParseReportHeaders(t *testing.T) {
	dir := t.TempDir()
	reportsDir := filepath.Join(dir, "reports")
	if err := os.MkdirAll(reportsDir, 0755); err != nil {
		t.Fatal(err)
	}
	content := `# 194 — Acme | Dev

**Score:** 4/5
**URL:** https://example.com/vaga
**Archetype:** Scaleup
**Legitimacy:** Alta

## Veredito

Vale a pena aplicar.
`
	if err := os.WriteFile(filepath.Join(reportsDir, "194-acme-2024-01-01.md"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	reports, err := New(dir).ParseReports()
	if err != nil {
		t.Fatalf("ParseReports erro: %v", err)
	}
	if len(reports) != 1 {
		t.Fatalf("got %d reports, want 1: %+v", len(reports), reports)
	}
	r := reports[0]
	if r.Num != 194 || r.Company != "Acme" || r.Role != "Dev" {
		t.Errorf("report = %+v", r)
	}
	if r.Score != "4/5" {
		t.Errorf("Score = %q, want 4/5", r.Score)
	}
	if r.Veredito != "Vale a pena aplicar." {
		t.Errorf("Veredito = %q", r.Veredito)
	}
	if r.Archetype != "Scaleup" || r.Legitimacy != "Alta" {
		t.Errorf("Archetype/Legitimacy = %q/%q", r.Archetype, r.Legitimacy)
	}
	if r.Date != "2024-01-01" {
		t.Errorf("Date = %q, want 2024-01-01 (extraído do filename)", r.Date)
	}
}
