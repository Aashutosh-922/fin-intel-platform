package clients

import (
	"context"

	"github.com/Aashutosh-922/fin-intel-platform/cmd/api-gateway/internal/handlers"
)

type AIClient struct{}

func NewAIClient() *AIClient {
	return &AIClient{}
}

func (c *AIClient) Query(
	ctx context.Context,
	q handlers.AIQuery,
) (handlers.AIResponse, error) {

	return handlers.AIResponse{
		Text: "AI explanation placeholder",
	}, nil
}