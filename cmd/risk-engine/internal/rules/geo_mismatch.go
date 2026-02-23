package rules

import "github.com/Aashutosh-922/fin-intel-platform/cmd/risk-engine/internal/domain/risk"

type GeoMismatchRule struct{}

func (r GeoMismatchRule) Name() string {
	return "geo_mismatch"
}

func (r GeoMismatchRule) Evaluate(ctx risk.Context) (risk.Factor, bool) {
	if ctx.LastCountry != "" && ctx.Country != ctx.LastCountry {
		return risk.Factor{
			Name:   r.Name(),
			Score:  20,
			Reason: "Transaction originated from a new country",
		}, true
	}
	return risk.Factor{}, false
}