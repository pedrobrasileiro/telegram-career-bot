package domain

type PipelineItem struct {
	URL     string `json:"url"`
	Company string `json:"company"`
	Title   string `json:"title"`
}

type Pipeline struct {
	Pending    []PipelineItem `json:"pending"`
	Processed  []PipelineItem `json:"processed"`
	ExportedAt string         `json:"exportedAt,omitempty"`
}
