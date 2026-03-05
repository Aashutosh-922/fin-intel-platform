package features

import "math"

func BehavioralDeviation(amount, mean, std float64) float64 {
	if std == 0 {
		return 0
	}
	z := (amount - mean) / std
	return math.Abs(z) / 5.0 // normalize
}