package application

import (
	"math"
	"time"

	"github.com/Aashutosh-922/fin-intel-platform/cmd/volatility-ai/internal/domain"
)

type Detector struct {
	windowSize int
	prices     map[string][]float64
}

func NewDetector(window int) *Detector {
	return &Detector{
		windowSize: window,
		prices:     make(map[string][]float64),
	}
}

func (d *Detector) AddTrade(symbol string, price float64) *domain.Alert {

	history := d.prices[symbol]
	history = append(history, price)

	if len(history) < d.windowSize {
		d.prices[symbol] = history
		return nil
	}

	if len(history) > d.windowSize {
		history = history[1:]
	}

	d.prices[symbol] = history

	vol := computeVolatility(history)
	mean := computeMean(history)
	std := computeStdDev(history)

	if std == 0 {
		return nil
	}

	z := (price - mean) / std

	if math.Abs(z) > 3 {
		return &domain.Alert{
			Symbol:     symbol,
			Type:       "VOLATILITY_SPIKE",
			ZScore:     z,
			Volatility: vol,
			Timestamp:  time.Now().Unix(),
		}
	}

	return nil
}

func computeMean(values []float64) float64 {
	var sum float64
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

func computeStdDev(values []float64) float64 {
	mean := computeMean(values)

	var variance float64
	for _, v := range values {
		diff := v - mean
		variance += diff * diff
	}

	return math.Sqrt(variance / float64(len(values)))
}

func computeVolatility(prices []float64) float64 {

	if len(prices) < 2 {
		return 0
	}

	var returns []float64

	for i := 1; i < len(prices); i++ {
		r := math.Log(prices[i] / prices[i-1])
		returns = append(returns, r)
	}

	return computeStdDev(returns)
}