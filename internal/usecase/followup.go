package usecase

import (
	"sort"
	"time"

	"telegram-career-bot/internal/domain"
	"telegram-career-bot/internal/port"
)

type FollowUpResult struct {
	Upcoming []domain.FollowUpItem // dentro dos próximos 7 dias, ordenado por data asc
}

type FollowUpQuery struct {
	Store port.SnapshotStore
}

func (q FollowUpQuery) Run(now time.Time) (*FollowUpResult, error) {
	data, err := q.Store.LoadFollowUps()
	if err != nil {
		return nil, err
	}

	nextWeek := now.Add(7 * 24 * time.Hour)

	var upcoming []domain.FollowUpItem
	for _, item := range data.Items {
		d, err := time.Parse("2006-01-02", item.Date)
		if err != nil {
			continue
		}
		if d.After(now) && d.Before(nextWeek) {
			upcoming = append(upcoming, item)
		}
	}

	sort.Slice(upcoming, func(i, j int) bool {
		return upcoming[i].Date < upcoming[j].Date
	})

	return &FollowUpResult{Upcoming: upcoming}, nil
}
