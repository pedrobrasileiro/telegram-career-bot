package usecase

// StatusOrder define a ordem de prioridade em que os status de aplicação
// são exibidos (tracker, stats, contexto pra /ask).
func StatusOrder() []string {
	return []string{"Interview", "Applied", "Offer", "Evaluated", "Responded", "Rejected", "Discarded", "SKIP"}
}
