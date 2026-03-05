package domain

type RiskDecisionEvent struct {
	EventID       string  `json:"event_id"`
	TransactionID string  `json:"transaction_id"`
	RiskScore     float64 `json:"risk_score"`
	Flagged       bool    `json:"flagged"`
	CreatedAt     int64   `json:"created_at"`
}