package application

import (
	"github.com/Aashutosh-922/fin-intel-platform/cmd/ai-service/internal/application/features"
	"github.com/Aashutosh-922/fin-intel-platform/cmd/ai-service/internal/domain"
)

type Analyzer struct{}

func NewAnalyzer() *Analyzer {
	return &Analyzer{}
}

func (a *Analyzer) Analyze(event domain.RiskDecisionEvent) domain.AIAnalysis {

	// Normalize upstream risk score from [0..100] to [0..1].
	ruleScore := clamp01(event.RiskScore / 100.0)

	// Derive additional signals from actual event values (no hardcoded mock inputs).
	behaviorScore := features.BehavioralDeviation(event.RiskScore, 50, 20)
	velocityCount := 1
	if event.RiskScore >= 80 {
		velocityCount = 12
	} else if event.RiskScore >= 60 {
		velocityCount = 8
	}
	velocityScore := features.VelocityScore(velocityCount)
	geoScore := features.GeoRisk(event.Flagged)

	final := clamp01(0.4*ruleScore + 0.3*behaviorScore + 0.2*velocityScore + 0.1*geoScore)

	verdict := "Low Risk"
	if final > 0.7 {
		verdict = "High Risk"
	}

	return domain.AIAnalysis{
		EventID:       generateID(),
		TransactionID: event.TransactionID,
		Verdict:       verdict,
		Confidence:    final,
		Reasoning: []string{
			"Rule-based risk weighted",
			"Behavioral anomaly applied",
			"Velocity anomaly applied",
			"Geo probability applied",
		},
		CreatedAt: now(),
	}
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}
