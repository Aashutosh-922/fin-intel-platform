package kafka

import (
	"context"
	"encoding/json"
	"log"

	"github.com/Aashutosh-922/fin-intel-platform/cmd/portfolio-service/internal/application"
	"github.com/Aashutosh-922/fin-intel-platform/cmd/portfolio-service/internal/domain"
	"github.com/segmentio/kafka-go"
)

func StartMarketConsumer(
	ctx context.Context,
	brokers []string,
	service *application.Service,
) {

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: brokers,
		GroupID: "portfolio-service-group",
		Topic:   "market-ticks",
	})

	for {
		m, err := reader.FetchMessage(ctx)
		if err != nil {
			continue
		}

		var tick domain.MarketTick
		if err := json.Unmarshal(m.Value, &tick); err != nil {
			continue
		}

		if err := service.ProcessMarketTick(ctx, tick); err != nil {
			log.Println(err)
			continue
		}

		reader.CommitMessages(ctx, m)
	}
}
