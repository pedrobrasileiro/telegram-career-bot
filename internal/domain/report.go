package domain

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

type ReportsIndexData struct {
	Reports    []ReportSummary `json:"reports"`
	ExportedAt string          `json:"exportedAt"`
}
