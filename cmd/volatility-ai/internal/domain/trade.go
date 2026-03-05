package domain

type Trade struct {
	Symbol    string  `json:"symbol"`
	Price     float64 `json:"price"`
	Quantity  int     `json:"quantity"`
	Timestamp int64   `json:"timestamp"`
}