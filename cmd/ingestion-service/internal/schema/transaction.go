package schema

type TransactionEvent struct {
	SchemaVersion string  `json:"schema_version"`
	EventID       string  `json:"event_id"`
	TransactionID string  `json:"transaction_id"`
	UserID        string  `json:"user_id"`
	Amount        float64 `json:"amount"`
	Currency      string  `json:"currency"`
	Country       string  `json:"country"`
	CreatedAt     string  `json:"created_at"`
}