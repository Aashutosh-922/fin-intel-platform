package domain

type Level struct {
	Price    float64 `json:"price"`
	Quantity int     `json:"quantity"`
}

type OrderBookSnapshot struct {
	Symbol string  `json:"symbol"`
	Bids   []Level `json:"bids"`
	Asks   []Level `json:"asks"`
	Spread float64 `json:"spread"`
}