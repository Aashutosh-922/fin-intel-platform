package handlers

type CreateTransactionRequest struct {
	UserID   string  `json:"user_id"`
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
}

type AIQuery struct {
	TransactionID string `json:"transaction_id"`
	Question      string `json:"question"`
}

type AIResponse struct {
	Text string `json:"text"`
}

type Transaction struct {
	ID     string
	Amount float64
	Status string
}

type Risk struct {
	Score float64
}

type AIExplanation struct {
	Text string
}

type Event struct {
	Type string
	Data string
}