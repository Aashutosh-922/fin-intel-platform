package repository

type TimelineRepository interface {
	GetTransactionHistory(transactionID string) ([]Event, error)
}

type Event struct {
	EventType string         `json:"event_type"`
	Payload   map[string]any `json:"payload"`
	CreatedAt int64          `json:"created_at"`
}
