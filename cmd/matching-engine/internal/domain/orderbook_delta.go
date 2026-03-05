package domain

type OrderBookDelta struct {
	Symbol    string  `json:"symbol"`
	Bids      []Level `json:"bids"`
	Asks      []Level `json:"asks"`
	Spread    float64 `json:"spread"`
	Timestamp int64   `json:"timestamp"`
}
