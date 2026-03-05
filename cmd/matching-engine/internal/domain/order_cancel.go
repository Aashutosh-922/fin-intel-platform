package domain

type OrderCancel struct {
	OrderID string `json:"order_id"`
	Symbol  string `json:"symbol"`
}