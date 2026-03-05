package domain

type OrderType string

const (
	Limit        OrderType = "LIMIT"
	Market       OrderType = "MARKET"
	StopLoss     OrderType = "STOP_LOSS"
	TakeProfit   OrderType = "TAKE_PROFIT"
	TrailingStop OrderType = "TRAILING_STOP"
)

type Order struct {
	OrderID      string    `json:"order_id"`
	UserID       string    `json:"user_id"`
	Symbol       string    `json:"symbol"`
	Side         string    `json:"side"` // BUY / SELL
	Type         OrderType `json:"type"`
	Price        float64   `json:"price"`
	Quantity     int       `json:"quantity"`
	Timestamp    int64     `json:"created_at"`
	Cancelled    bool      `json:"-"`
	StopPrice    float64   `json:"stop_price,omitempty"`
	TriggerPrice float64   `json:"trigger_price,omitempty"`
	TrailAmount  float64   `json:"trail_amount,omitempty"`
	HighestSeen  float64   `json:"-"`
	LowestSeen   float64   `json:"-"`
}
