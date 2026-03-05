package domain

type OrderSide string
type OrderType string

const (
	Buy  OrderSide = "BUY"
	Sell OrderSide = "SELL"

	Limit  OrderType = "LIMIT"
	Market OrderType = "MARKET"
	StopLoss  OrderType = "STOP_LOSS"
	TakeProfit    OrderType = "TAKE_PROFIT"
	TrailingStop  OrderType = "TRAILING_STOP"
)

type Order struct {
	OrderID   string
	UserID    string
	Symbol    string
	Side      string
	Type      OrderType
	Price     float64
	Quantity  int
	Timestamp int64
	Cancelled bool
	StopPrice float64
	TriggerPrice float64   // for take profit
    TrailAmount float64   // for trailing stop
    HighestSeen float64   // runtime tracking
    LowestSeen  float64   // runtime tracking
}