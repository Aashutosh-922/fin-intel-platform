package kafka

import (
	"context"
	"encoding/json"

	"github.com/Aashutosh-922/fin-intel-platform/cmd/market-data-service/internal/domain"
	"github.com/segmentio/kafka-go"
)

type Producer struct {
	writer *kafka.Writer
}

func NewProducer(brokers []string, topic string) *Producer {
	return &Producer{
		writer: &kafka.Writer{
			Addr:         kafka.TCP(brokers...),
			Topic:        topic,
			RequiredAcks: kafka.RequireAll,
		},
	}
}

func (p *Producer) Publish(ctx context.Context, tick domain.MarketTick) error {

	payload, err := json.Marshal(tick)
	if err != nil {
		return err
	}

	msg := kafka.Message{
		Key:   []byte(tick.Symbol), // 🔥 partition by symbol
		Value: payload,
	}

	return p.writer.WriteMessages(ctx, msg)
}

func (p *Producer) Close() error {
	return p.writer.Close()
}
