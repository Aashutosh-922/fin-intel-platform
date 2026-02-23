package rules

import "github.com/Aashutosh-922/fin-intel-platform/cmd/risk-engine/internal/domain/risk"

type AmountSpikeRule struct{}

func (r AmountSpikeRule) Name() string {
	return "amount_spike"
}

func (r AmountSpikeRule) Evaluate(ctx risk.Context) (risk.Factor, bool) {
	if ctx.AvgAmountLast30d > 0 && ctx.Amount > ctx.AvgAmountLast30d*3 {
		return risk.Factor{
			Name:   r.Name(),
			Score:  30,
			Reason: "Transaction amount significantly higher than historical average",
		}, true
	}
	return risk.Factor{}, false
}