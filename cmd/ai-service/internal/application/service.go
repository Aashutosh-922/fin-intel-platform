package application

import (
	"context"
	"errors"
	"strings"

	"github.com/Aashutosh-922/fin-intel-platform/cmd/ai-service/internal/domain"
)

type LLMClient interface {
	Analyze(ctx context.Context, event domain.RiskDecisionEvent) (domain.AIAnalysis, error)
}

type Service struct {
	analyzer *Analyzer
	llm      LLMClient
	mode     string
}

func NewService(analyzer *Analyzer, llm LLMClient, mode string) *Service {
	if mode == "" {
		mode = "deterministic"
	}
	return &Service{
		analyzer: analyzer,
		llm:      llm,
		mode:     strings.ToLower(strings.TrimSpace(mode)),
	}
}

func (s *Service) Process(ctx context.Context, event domain.RiskDecisionEvent) (domain.AIAnalysis, error) {
	fallback := s.analyzer.Analyze(event)

	switch s.mode {
	case "deterministic":
		return fallback, nil
	case "llm_only":
		if s.llm == nil {
			return domain.AIAnalysis{}, errors.New("llm_only mode but no llm client configured")
		}
		return s.llm.Analyze(ctx, event)
	case "hybrid":
		if s.llm == nil {
			return fallback, nil
		}
		analysis, err := s.llm.Analyze(ctx, event)
		if err != nil {
			return fallback, nil
		}
		return analysis, nil
	default:
		return fallback, nil
	}
}
