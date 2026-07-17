package domain

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
