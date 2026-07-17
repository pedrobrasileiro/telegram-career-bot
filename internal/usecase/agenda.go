package usecase

import (
	"sort"

	"telegram-career-bot/internal/domain"
	"telegram-career-bot/internal/port"
)

type AgendaQuery struct {
	Store port.SnapshotStore
}

func (q AgendaQuery) Run() ([]domain.Application, error) {
	tracker, err := q.Store.LoadTracker()
	if err != nil {
		return nil, err
	}

	var interviews []domain.Application
	for _, a := range tracker.Applications {
		if a.Status == "Interview" {
			interviews = append(interviews, a)
		}
	}

	sort.Slice(interviews, func(i, j int) bool {
		return interviews[i].Num > interviews[j].Num
	})

	return interviews, nil
}
