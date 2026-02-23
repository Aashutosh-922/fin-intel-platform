package handlers

import "context"

type MockAIClient struct{}

func NewMockAIClient() *MockAIClient {
	return &MockAIClient{}
}

func (m *MockAIClient) Query(ctx context.Context, q string) (AIResponse, error) {
	return AIResponse{
		Text: "Mock AI response",
	}, nil
}