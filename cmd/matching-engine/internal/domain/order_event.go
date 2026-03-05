package domain

type OrderEvent struct {
	OrderID string `json:"order_id"`
	Type    string `json:"type"`
}