package risk

import "time"

type Context struct {
	TransactionID string
	UserID        string
	Amount        float64
	Country       string
	CreatedAt     time.Time

	AvgAmountLast30d float64
	TxCountLast1h    int
	LastCountry      string
}