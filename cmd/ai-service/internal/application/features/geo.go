package features

func GeoRisk(crossBorder bool) float64 {
	if crossBorder {
		return 0.7
	}
	return 0.2
}
