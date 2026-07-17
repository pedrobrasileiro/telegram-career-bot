package usecase

import (
	"strconv"
	"strings"

	"telegram-career-bot/internal/domain"
	"telegram-career-bot/internal/port"
)

// ComputeStats agrega os totais/funil a partir da lista de applications
// (usado pelo Export ao gerar stats.json).
func ComputeStats(apps []domain.Application) domain.Stats {
	s := domain.Stats{ByStatus: make(map[string]int)}

	var totalScore float64
	var scoreCount int

	for _, a := range apps {
		s.Total++
		s.ByStatus[a.Status]++

		scoreStr := strings.TrimSuffix(a.Score, "/5")
		if score, err := strconv.ParseFloat(scoreStr, 64); err == nil {
			totalScore += score
			scoreCount++
		}
	}

	if scoreCount > 0 {
		s.AverageScore = totalScore / float64(scoreCount)
	}

	s.Funnel = domain.Funnel{
		Total:     s.Total,
		Evaluated: s.ByStatus["Evaluated"],
		Applied:   s.ByStatus["Applied"],
		Interview: s.ByStatus["Interview"],
		Offer:     s.ByStatus["Offer"],
		Rejected:  s.ByStatus["Rejected"],
		Discarded: s.ByStatus["Discarded"],
		Skipped:   s.ByStatus["SKIP"],
	}

	return s
}

// StatsResult é o funil completo (aplicadas → resposta → entrevista → oferta)
// derivado do snapshot de stats, pronto pra apresentação.
type StatsResult struct {
	Stats         domain.Stats
	EverApplied   int
	EverResponded int
	Interview     int
	Offer         int
	Rejected      int
	Active        int
	Skipped       int
	ResponseRate  float64
	InterviewRate float64
}

type StatsQuery struct {
	Store port.SnapshotStore
}

func (q StatsQuery) Run() (*StatsResult, error) {
	stats, err := q.Store.LoadStats()
	if err != nil {
		return nil, err
	}
	if _, err := q.Store.LoadTracker(); err != nil {
		return nil, err
	}

	byStatus := stats.ByStatus

	everApplied := byStatus["Applied"] + byStatus["Interview"] + byStatus["Offer"] + byStatus["Rejected"] + byStatus["Responded"]
	everResponded := byStatus["Interview"] + byStatus["Offer"] + byStatus["Rejected"] + byStatus["Responded"]
	interview := byStatus["Interview"] + byStatus["Offer"]
	offer := byStatus["Offer"]
	rejected := byStatus["Rejected"]
	active := byStatus["Applied"] + byStatus["Interview"] + byStatus["Responded"] + byStatus["Evaluated"]
	skipped := byStatus["Discarded"] + byStatus["SKIP"]

	respRate := 0.0
	if everApplied > 0 {
		respRate = float64(everResponded) / float64(everApplied) * 100
	}
	interviewRate := 0.0
	if everApplied > 0 {
		interviewRate = float64(interview) / float64(everApplied) * 100
	}

	return &StatsResult{
		Stats:         *stats,
		EverApplied:   everApplied,
		EverResponded: everResponded,
		Interview:     interview,
		Offer:         offer,
		Rejected:      rejected,
		Active:        active,
		Skipped:       skipped,
		ResponseRate:  respRate,
		InterviewRate: interviewRate,
	}, nil
}
