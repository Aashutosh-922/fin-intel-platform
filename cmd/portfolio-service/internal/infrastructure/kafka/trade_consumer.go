package kafka

import (
	"context"
	"encoding/json"
	"log"

	"github.com/segmentio/kafka-go"
	"github.com/Aashutosh-922/fin-intel-platform/cmd/portfolio-service/internal/application"
	"github.com/Aashutosh-922/fin-intel-platform/cmd/portfolio-service/internal/domain"
)

func StartTradeConsumer(
	ctx context.Context,
	brokers []string,
	service *application.Service,
) {

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: brokers,
		GroupID: "portfolio-service-group",
		Topic:   "trade-executed",
	})

	for {
		m, err := reader.FetchMessage(ctx)
		if err != nil {
			continue
		}

		var trade domain.Trade
		if err := json.Unmarshal(m.Value, &trade); err != nil {
			continue
		}

		if err := service.ProcessTrade(ctx, trade); err != nil {
			log.Println(err)
			continue
		}

		reader.CommitMessages(ctx, m)
	}
}