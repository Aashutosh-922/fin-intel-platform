package application

import (
	"math/rand"
	"time"
    "github.com/Aashutosh-922/fin-intel-platform/cmd/market-data-service/internal/domain"
)

type Generator struct {
	prices map[string]float64
}

func NewGenerator(symbols []string) *Generator {
	rand.Seed(time.Now().UnixNano())

	prices := make(map[string]float64)
	for _, s := range symbols {
		prices[s] = 1000 + rand.Float64()*100
	}

	return &Generator{prices: prices}
}

func (g *Generator) NextTick(symbol string) domain.MarketTick {
	current := g.prices[symbol]

	// Random walk
	change := rand.Float64()*2 - 1
	newPrice := current + change

	if newPrice < 1 {
		newPrice = 1
	}

	g.prices[symbol] = newPrice

	return domain.MarketTick{
		Symbol:    symbol,
		Price:     newPrice,
		Timestamp: time.Now().Unix(),
	}
}