package domain

type Position struct {
	UserID        string
	Symbol        string
	Quantity      int
	AvgPrice      float64
	RealizedPnL   float64
	UnrealizedPnL float64
}