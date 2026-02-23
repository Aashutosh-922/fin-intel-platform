package events

import "time"

type TransactionEvent struct {
	Version    string    `json:"version"`
	EventID    string    `json:"event_id"`
	UserID     string    `json:"user_id"`
	Amount     float64   `json:"amount"`
	Currency   string    `json:"currency"`
	OccurredAt time.Time `json:"occurred_at"`
	Source     string    `json:"source"`
}
