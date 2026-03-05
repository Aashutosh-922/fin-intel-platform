package domain

type Alert struct {
	Symbol     string  `json:"symbol"`
	Type       string  `json:"type"`
	ZScore     float64 `json:"z_score"`
	Volatility float64 `json:"volatility"`
	Timestamp  int64   `json:"timestamp"`
}