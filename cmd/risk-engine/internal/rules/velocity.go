package rules

import "github.com/Aashutosh-922/fin-intel-platform/cmd/risk-engine/internal/domain/risk"

type VelocityRule struct{}

func (r VelocityRule) Name() string {
	return "velocity"
}

func (r VelocityRule) Evaluate(ctx risk.Context) (risk.Factor, bool) {
	if ctx.TxCountLast1h >= 5 {
		return risk.Factor{
			Name:   r.Name(),
			Score:  25,
			Reason: "High transaction frequency detected",
		}, true
	}
	return risk.Factor{}, false
}