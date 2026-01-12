package events

import (
	"context"
	"time"
)

type Event struct {
	TransactionID string
	Type          string
	Time          time.Time
	Metadata      map[string]interface{}
}

type Repository interface {
	Insert(ctx context.Context, e Event) error
}

type Service struct {
	repo Repository
}

func New(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Record(ctx context.Context, e Event) error {
	return s.repo.Insert(ctx, e)
}
