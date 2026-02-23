package producer

import (
	"context"
	"encoding/json"

	"github.com/twmb/franz-go/pkg/kgo"
)

const topic = "risk-decisions"

type Producer struct {
	client *kgo.Client
}

func New(client *kgo.Client) *Producer {
	return &Producer{client: client}
}

type RiskEvent struct {
	TransactionID string `json:"transaction_id"`
	RiskScore     int    `json:"risk_score"`
	Decision      string `json:"decision"`
}

func (p *Producer) Publish(ctx context.Context, evt RiskEvent) error {
	bytes, err := json.Marshal(evt)
	if err != nil {
		return err
	}

	rec := &kgo.Record{
		Topic: topic,
		Value: bytes,
		Key:   []byte(evt.TransactionID),
	}

	return p.client.ProduceSync(ctx, rec).FirstErr()
}