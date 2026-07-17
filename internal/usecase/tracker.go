package usecase

import (
	"sort"
	"strings"

	"telegram-career-bot/internal/domain"
	"telegram-career-bot/internal/port"
)

type TrackerResult struct {
	Total        int
	StatusCounts map[string]int
	Filter       string
	Filtered     []domain.Application // ordenado desc por Num
}

type TrackerQuery struct {
	Store port.SnapshotStore
}

func (q TrackerQuery) Run(filter string) (*TrackerResult, error) {
	tracker, err := q.Store.LoadTracker()
	if err != nil {
		return nil, err
	}

	statusCounts := make(map[string]int)
	for _, a := range tracker.Applications {
		statusCounts[a.Status]++
	}

	var filtered []domain.Application
	if filter != "" {
		for _, a := range tracker.Applications {
			if strings.EqualFold(a.Status, filter) {
				filtered = append(filtered, a)
			}
		}
	} else {
		filtered = tracker.Applications
	}

	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].Num > filtered[j].Num
	})

	return &TrackerResult{
		Total:        len(tracker.Applications),
		StatusCounts: statusCounts,
		Filter:       filter,
		Filtered:     filtered,
	}, nil
}
