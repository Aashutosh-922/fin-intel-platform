package transaction

import "time"

type Transaction struct {
	ID        string
	UserID    string
	Amount    float64
	Currency  string
	Country   string
	Status    string
	CreatedAt time.Time
}
