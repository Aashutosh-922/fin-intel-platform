package kafka

import (
	"context"
	"encoding/json"

	"github.com/segmentio/kafka-go"
	"github.com/Aashutosh-922/fin-intel-platform/cmd/volatility-ai/internal/domain"
)

type Producer struct {
	writer *kafka.Writer
}

func NewProducer(brokers []string, topic string) *Producer {
	return &Producer{
		writer: &kafka.Writer{
			Addr:  kafka.TCP(brokers...),
			Topic: topic,
		},
	}
}

func (p *Producer) Publish(ctx context.Context, alert domain.Alert) error {

	payload, err := json.Marshal(alert)
	if err != nil {
		return err
	}

	return p.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(alert.Symbol),
		Value: payload,
	})
}