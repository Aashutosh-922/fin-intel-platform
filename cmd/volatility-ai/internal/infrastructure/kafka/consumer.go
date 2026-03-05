package kafka

import (
	"context"
	"encoding/json"
	"log"

	"github.com/segmentio/kafka-go"
	"github.com/Aashutosh-922/fin-intel-platform/cmd/volatility-ai/internal/application"
	"github.com/Aashutosh-922/fin-intel-platform/cmd/volatility-ai/internal/domain"
)

type Consumer struct {
	reader   *kafka.Reader
	service  *application.Service
	producer *Producer
}

func NewConsumer(
	brokers []string,
	groupID string,
	topic string,
	service *application.Service,
	producer *Producer,
) *Consumer {

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: brokers,
		GroupID: groupID,
		Topic:   topic,
	})

	return &Consumer{
		reader:   reader,
		service:  service,
		producer: producer,
	}
}

func (c *Consumer) Start(ctx context.Context) {

	for {
		msg, err := c.reader.FetchMessage(ctx)
		if err != nil {
			log.Println("fetch error:", err)
			continue
		}

		var trade domain.Trade
		if err := json.Unmarshal(msg.Value, &trade); err != nil {
			log.Println("unmarshal error:", err)
			continue
		}

		alert := c.service.ProcessTrade(trade)

		if alert != nil {
			if err := c.producer.Publish(ctx, *alert); err != nil {
				log.Println("publish error:", err)
			}
		}

		c.reader.CommitMessages(ctx, msg)
	}
}