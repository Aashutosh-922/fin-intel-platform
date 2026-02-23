package repository

import (
	"github.com/Aashutosh-922/fin-intel-platform/cmd/api-gateway/internal/handlers"
)

type ReadOnlyRepo struct{}

func NewReadOnlyRepo() *ReadOnlyRepo {
	return &ReadOnlyRepo{}
}

func (r *ReadOnlyRepo) GetTransaction(id string) (handlers.Transaction, error) {
	return handlers.Transaction{}, nil
}

func (r *ReadOnlyRepo) GetRisk(id string) (handlers.Risk, error) {
	return handlers.Risk{}, nil
}

func (r *ReadOnlyRepo) GetExplanation(id string) (handlers.AIExplanation, error) {
	return handlers.AIExplanation{}, nil
}

// 🔴 THIS WAS MISSING — now fixed
func (r *ReadOnlyRepo) GetEvents(id string) ([]handlers.Event, error) {
	return []handlers.Event{}, nil
}
