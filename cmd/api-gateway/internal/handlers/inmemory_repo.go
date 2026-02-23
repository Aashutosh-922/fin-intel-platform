package handlers

type InMemoryRepo struct{}

func NewInMemoryRepo() *InMemoryRepo {
	return &InMemoryRepo{}
}

func (r *InMemoryRepo) GetTransaction(id string) (Transaction, error) {
	return Transaction{ID: id, Amount: 1200, Status: "PENDING"}, nil
}

func (r *InMemoryRepo) GetRisk(id string) (Risk, error) {
	return Risk{Score: 42}, nil
}

func (r *InMemoryRepo) GetExplanation(id string) (AIExplanation, error) {
	return AIExplanation{Text: "Mock explanation"}, nil
}

func (r *InMemoryRepo) GetEvents(id string) ([]Event, error) {
	return []Event{}, nil
}