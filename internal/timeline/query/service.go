// package query

// import "context"

// type Service struct {
// 	repo *Repository
// }

// func NewService(repo *Repository) *Service {
// 	return &Service{repo: repo}
// }

// func (s *Service) Timeline(ctx context.Context, txnID string) ([]Event, error) {
// 	return s.repo.GetTimeline(ctx, txnID)
// }

package query

import "context"

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetTimeline(ctx context.Context, txnID string) (Timeline, error) {
	events, err := s.repo.GetTimeline(ctx, txnID)
	if err != nil {
		return Timeline{}, err
	}

	return Timeline{
		TransactionID: txnID,
		Events:        events,
	}, nil
}