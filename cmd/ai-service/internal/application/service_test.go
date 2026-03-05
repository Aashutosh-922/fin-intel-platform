package application

import (
	"context"
	"errors"
	"testing"

	"github.com/Aashutosh-922/fin-intel-platform/cmd/ai-service/internal/domain"
)

type fakeLLM struct {
	output domain.AIAnalysis
	err    error
}

func (f fakeLLM) Analyze(_ context.Context, _ domain.RiskDecisionEvent) (domain.AIAnalysis, error) {
	if f.err != nil {
		return domain.AIAnalysis{}, f.err
	}
	return f.output, nil
}

func TestProcessHybridFallsBackWhenLLMErrors(t *testing.T) {
	svc := NewService(NewAnalyzer(), fakeLLM{err: errors.New("boom")}, "hybrid")
	event := domain.RiskDecisionEvent{TransactionID: "tx-1", RiskScore: 90, Flagged: true}

	got, err := svc.Process(context.Background(), event)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.TransactionID != event.TransactionID {
		t.Fatalf("expected fallback analysis for transaction %s, got %s", event.TransactionID, got.TransactionID)
	}
}

func TestProcessLLMOnlyRequiresLLMClient(t *testing.T) {
	svc := NewService(NewAnalyzer(), nil, "llm_only")

	_, err := svc.Process(context.Background(), domain.RiskDecisionEvent{TransactionID: "tx-1"})
	if err == nil {
		t.Fatalf("expected error when llm_only mode has no llm client")
	}
}

func TestProcessLLMOnlyUsesLLMOutput(t *testing.T) {
	expected := domain.AIAnalysis{
		EventID:       "llm-1",
		TransactionID: "tx-llm",
		Verdict:       "High Risk",
		Confidence:    0.88,
		Reasoning:     []string{"model"},
		CreatedAt:     1,
	}
	svc := NewService(NewAnalyzer(), fakeLLM{output: expected}, "llm_only")

	got, err := svc.Process(context.Background(), domain.RiskDecisionEvent{TransactionID: "tx-llm"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Verdict != expected.Verdict || got.Confidence != expected.Confidence {
		t.Fatalf("expected llm output %#v, got %#v", expected, got)
	}
}
