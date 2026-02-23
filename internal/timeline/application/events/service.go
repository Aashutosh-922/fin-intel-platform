// package events

// import (
// 	"context"
// 	"time"

// 	"github.com/google/uuid"
// )

// type Repository interface {
// 	Insert(ctx context.Context, e Event) error
// }

// type Service struct {
// 	repo Repository
// }

// func New(repo Repository) *Service {
// 	return &Service{repo: repo}
// }

// func (s *Service) Record(ctx context.Context, e Event) error {
// 	e.ID = uuid.NewString()          // unique timeline event id
// 	e.CreatedAt = time.Now().UTC()   // authoritative time

// 	return s.repo.Insert(ctx, e)
// }

package events

import (
	"context"
	"time"

	"github.com/google/uuid"
)

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
	// authoritative event enrichment
	e.ID = uuid.NewString()
	e.CreatedAt = time.Now().UTC()

	return s.repo.Insert(ctx, e)
}