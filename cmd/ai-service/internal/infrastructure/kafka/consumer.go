package kafka

import (
	"context"
	"encoding/json"
	"log"

	"github.com/Aashutosh-922/fin-intel-platform/cmd/ai-service/internal/application"
	"github.com/Aashutosh-922/fin-intel-platform/cmd/ai-service/internal/domain"
	"github.com/Aashutosh-922/fin-intel-platform/internal/contracts"
	"github.com/segmentio/kafka-go"
)

type Consumer struct {
	reader   *kafka.Reader
	service  *application.Service
	producer *Producer
}

func NewConsumer(brokers []string, groupID, topic string, service *application.Service, producer *Producer) *Consumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: brokers,
		GroupID: groupID,
		Topic:   topic,
	})
	return &Consumer{reader: reader, service: service, producer: producer}
}

func (c *Consumer) Start(ctx context.Context) {
	for {
		msg, err := c.reader.FetchMessage(ctx)
		if err != nil {
			log.Println("fetch error:", err)
			continue
		}

		var event domain.RiskDecisionEvent
		if err := contracts.ValidateTopicPayload("risk-decisions", msg.Value); err != nil {
			log.Println("contract validation failed:", err)
			continue
		}
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			log.Println("unmarshal error:", err)
			continue
		}

		analysis, err := c.service.Process(ctx, event)
		if err != nil {
			log.Println("analysis error:", err)
			continue
		}
		if err := c.producer.Publish(ctx, analysis); err != nil {
			log.Println("publish error:", err)
			continue
		}

		if err := c.reader.CommitMessages(ctx, msg); err != nil {
			log.Println("commit error:", err)
		}
	}
}
