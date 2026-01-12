package kafka

import (
	"context"
	"encoding/json"

	"github.com/Aashutosh-922/fin-intel-platform/internal/domain/transaction"
)

type Producer struct {
	// real producer client later
}

func NewProducer() *Producer {
	return &Producer{}
}

func (p *Producer) PublishTransactionReceived(
	ctx context.Context,
	tx transaction.Transaction,
) error {
	payload := map[string]interface{}{
		"event_version":  1,
		"transaction_id": tx.ID,
		"user_id":        tx.UserID,
		"amount":         tx.Amount,
		"currency":       tx.Currency,
		"country":        tx.Country,
		"created_at":     tx.CreatedAt,
	}

	_, _ = json.Marshal(payload)
	// send to Kafka here

	return nil
}
