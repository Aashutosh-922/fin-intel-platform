package query

import "time"

type TimelineEvent struct {
	EventType string    `json:"event"`
	CreatedAt time.Time `json:"time"`
}

type Timeline struct {
	TransactionID string          `json:"transaction_id"`
	Events        []TimelineEvent `json:"timeline"`
}