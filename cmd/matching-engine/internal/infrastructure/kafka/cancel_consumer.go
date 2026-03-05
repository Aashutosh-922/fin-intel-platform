package kafka

import (
	"context"
	"encoding/json"

	"github.com/Aashutosh-922/fin-intel-platform/cmd/matching-engine/internal/application"
	"github.com/Aashutosh-922/fin-intel-platform/cmd/matching-engine/internal/domain"
	"github.com/Aashutosh-922/fin-intel-platform/internal/contracts"
	"github.com/segmentio/kafka-go"
)

func StartCancelConsumer(
	ctx context.Context,
	brokers []string,
	matcher *application.Matcher,
) {

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: brokers,
		GroupID: "matching-engine-group",
		Topic:   "order-cancel",
	})

	for {
		m, err := reader.FetchMessage(ctx)
		if err != nil {
			continue
		}

		var cancel domain.OrderCancel
		if err := contracts.ValidateTopicPayload("order-cancel", m.Value); err != nil {
			continue
		}
		if err := json.Unmarshal(m.Value, &cancel); err != nil {
			continue
		}

		matcher.CancelOrder(cancel)

		reader.CommitMessages(ctx, m)
	}
}
