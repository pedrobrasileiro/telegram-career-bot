package usecase

import (
	"testing"

	"telegram-career-bot/internal/domain"
)

func TestComputeStats(t *testing.T) {
	apps := []domain.Application{
		{Status: "Applied", Score: "4/5"},
		{Status: "Applied", Score: "3/5"},
		{Status: "Interview", Score: "5/5"},
		{Status: "Rejected", Score: "N/A"},
	}
	stats := ComputeStats(apps)

	if stats.Total != 4 {
		t.Errorf("Total = %d, want 4", stats.Total)
	}
	if stats.ByStatus["Applied"] != 2 {
		t.Errorf("ByStatus[Applied] = %d, want 2", stats.ByStatus["Applied"])
	}
	if stats.ByStatus["Interview"] != 1 {
		t.Errorf("ByStatus[Interview] = %d, want 1", stats.ByStatus["Interview"])
	}
	wantAvg := (4.0 + 3.0 + 5.0) / 3.0
	if stats.AverageScore != wantAvg {
		t.Errorf("AverageScore = %v, want %v", stats.AverageScore, wantAvg)
	}
	if stats.Funnel.Applied != 2 || stats.Funnel.Interview != 1 || stats.Funnel.Rejected != 1 {
		t.Errorf("Funnel incorreto: %+v", stats.Funnel)
	}
}

func TestComputeStatsSemScores(t *testing.T) {
	stats := ComputeStats(nil)
	if stats.Total != 0 || stats.AverageScore != 0 {
		t.Errorf("stats de lista vazia deveria zerar tudo: %+v", stats)
	}
}

func TestStatusOrderContemTodosStatusConhecidos(t *testing.T) {
	order := StatusOrder()
	want := []string{"Interview", "Applied", "Offer", "Evaluated", "Responded", "Rejected", "Discarded", "SKIP"}
	if len(order) != len(want) {
		t.Fatalf("got %d status, want %d: %v", len(order), len(want), order)
	}
	for i, w := range want {
		if order[i] != w {
			t.Errorf("StatusOrder[%d] = %q, want %q", i, order[i], w)
		}
	}
}
