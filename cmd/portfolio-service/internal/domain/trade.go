package domain

type Trade struct {
	TradeID    string  `json:"trade_id"`
	BuyOrder   string  `json:"buy_order_id"`
	SellOrder  string  `json:"sell_order_id"`
	UserID     string  `json:"user_id,omitempty"` // backward compatibility
	BuyUserID  string  `json:"buy_user_id,omitempty"`
	SellUserID string  `json:"sell_user_id,omitempty"`
	Symbol     string  `json:"symbol"`
	Price      float64 `json:"price"`
	Quantity   int     `json:"quantity"`
	Timestamp  int64   `json:"timestamp"`
}
