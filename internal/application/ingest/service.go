package ingest

import (
	"context"
	"errors"

	"github.com/Aashutosh-922/fin-intel-platform/internal/domain/transaction"
)

type IdempotencyRepository interface {
	Get(ctx context.Context, key string) (string, bool, error)
	Save(ctx context.Context, key, transactionID string) error
}

type Repository interface {
	Save(ctx context.Context, tx transaction.Transaction) error
	GetByID(ctx context.Context, id string) (transaction.Transaction, error)
}

type EventPublisher interface {
	PublishTransactionReceived(ctx context.Context, tx transaction.Transaction) error
}

type Service struct {
	repo       Repository
	idempoRepo IdempotencyRepository
	pub        EventPublisher
}

func New(
	repo Repository,
	idRepo IdempotencyRepository,
	pub EventPublisher,
) *Service {
	return &Service{
		repo:       repo,
		idempoRepo: idRepo,
		pub:        pub,
	}
}

func (s *Service) Ingest(
	ctx context.Context,
	idempotencyKey string,
	tx transaction.Transaction,
) (transaction.Transaction, error) {

	if idempotencyKey == "" {
		return transaction.Transaction{}, errors.New("missing idempotency key")
	}

	// 1️⃣ Check existing
	existingTxID, found, err := s.idempoRepo.Get(ctx, idempotencyKey)
	if err != nil {
		return transaction.Transaction{}, err
	}

	if found {
		return s.repo.GetByID(ctx, existingTxID)
	}

	// 2️⃣ Save transaction
	if err := s.repo.Save(ctx, tx); err != nil {
		return transaction.Transaction{}, err
	}

	// 3️⃣ Save idempotency mapping
	if err := s.idempoRepo.Save(ctx, idempotencyKey, tx.ID); err != nil {
		return transaction.Transaction{}, err
	}

	// 4️⃣ Publish Kafka event
	if err := s.pub.PublishTransactionReceived(ctx, tx); err != nil {
		return transaction.Transaction{}, err
	}

	return tx, nil
}
