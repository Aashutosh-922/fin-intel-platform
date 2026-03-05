package domain

type OrderCreatedEvent struct {
	EventID   string  `json:"event_id"`
	OrderID   string  `json:"order_id"`
	UserID    string  `json:"user_id"`
	Symbol    string  `json:"symbol"`
	Side      string  `json:"side"`
	Price     float64 `json:"price"`
	Quantity  int     `json:"quantity"`
	Type      string  `json:"type"`
	CreatedAt int64   `json:"created_at"`
}