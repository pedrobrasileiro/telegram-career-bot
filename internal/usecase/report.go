package usecase

import (
	"telegram-career-bot/internal/domain"
	"telegram-career-bot/internal/port"
)

type ReportQuery struct {
	Store port.SnapshotStore
}

// Run retorna (nil, nil) se o snapshot existe mas o report não foi
// encontrado; retorna (nil, err) se o snapshot nem existe (sync pendente).
func (q ReportQuery) Run(num int) (*domain.ReportSummary, error) {
	data, err := q.Store.LoadReports()
	if err != nil {
		return nil, err
	}

	for _, r := range data.Reports {
		if r.Num == num {
			return &r, nil
		}
	}
	return nil, nil
}
