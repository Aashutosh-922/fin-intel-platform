package kafka

import (
	"context"
	"encoding/json"
	"log"

	"github.com/segmentio/kafka-go"
	"github.com/Aashutosh-922/fin-intel-platform/cmd/portfolio-service/internal/application"
	"github.com/Aashutosh-922/fin-intel-platform/cmd/portfolio-service/internal/domain"
)

type Consumer struct {
	reader  *kafka.Reader
	service *application.Service
}

func (c *Consumer) Start(ctx context.Context) {
	for {
		m, err := c.reader.FetchMessage(ctx)
		if err != nil {
			log.Println("error:", err)
			continue
		}

		var trade domain.Trade
		if err := json.Unmarshal(m.Value, &trade); err != nil {
			continue
		}

		if err := c.service.ProcessTrade(ctx, trade); err != nil {
			continue
		}

		c.reader.CommitMessages(ctx, m)
	}
}