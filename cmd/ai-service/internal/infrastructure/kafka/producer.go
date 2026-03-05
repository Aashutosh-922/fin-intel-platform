package kafka

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/Aashutosh-922/fin-intel-platform/cmd/ai-service/internal/domain"
	"github.com/segmentio/kafka-go"
)

type Producer struct {
	writer *kafka.Writer
}

func NewProducer(brokersCSV, topic string) *Producer {
	brokers := strings.Split(brokersCSV, ",")
	return &Producer{
		writer: &kafka.Writer{
			Addr:         kafka.TCP(brokers...),
			Topic:        topic,
			RequiredAcks: kafka.RequireAll,
		},
	}
}

func (p *Producer) Publish(ctx context.Context, analysis domain.AIAnalysis) error {
	payload, err := json.Marshal(analysis)
	if err != nil {
		return err
	}
	return p.writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(analysis.TransactionID),
		Value: payload,
	})
}

func (p *Producer) Close() error {
	return p.writer.Close()
}
