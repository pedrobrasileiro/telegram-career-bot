package domain

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

type TrackerData struct {
	Applications []Application `json:"applications"`
	ExportedAt   string        `json:"exportedAt"`
}
