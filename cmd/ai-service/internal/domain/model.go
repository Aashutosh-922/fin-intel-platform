package domain

type AIAnalysis struct {
	EventID       string   `json:"event_id"`
	TransactionID string   `json:"transaction_id"`
	Verdict       string   `json:"verdict"`
	Confidence    float64  `json:"confidence"`
	Reasoning     []string `json:"reasoning"`
	CreatedAt     int64    `json:"created_at"`
}
