package domain

import "time"

type Task struct {
	Type        string
	Description string
	StartTime   time.Time
}
