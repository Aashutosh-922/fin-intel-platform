package http

type CreateTransactionRequest struct {
	UserID   string  `json:"user_id"`
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
	Country  string  `json:"country"`
}
