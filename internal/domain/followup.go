package domain

type FollowUpItem struct {
	Date    string `json:"date"`
	Company string `json:"company"`
	Role    string `json:"role"`
	Action  string `json:"action"`
	Notes   string `json:"notes"`
}

type FollowUpData struct {
	Items      []FollowUpItem `json:"items"`
	ExportedAt string         `json:"exportedAt"`
}
