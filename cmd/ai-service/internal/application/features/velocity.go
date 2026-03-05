package features

func VelocityScore(txCountLastMinute int) float64 {
	if txCountLastMinute > 10 {
		return 0.8
	}
	return 0.2
}