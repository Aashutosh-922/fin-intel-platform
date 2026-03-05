package kafka

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/Aashutosh-922/fin-intel-platform/cmd/matching-engine/internal/application"
	"github.com/Aashutosh-922/fin-intel-platform/cmd/matching-engine/internal/domain"
	"github.com/Aashutosh-922/fin-intel-platform/internal/contracts"
	"github.com/segmentio/kafka-go"
)

type SnapshotStore interface {
	Save(symbol string, snapshot domain.OrderBookSnapshot) error
}

type Consumer struct {
	reader           *kafka.Reader
	matcher          *application.Matcher
	tradeProducer    *Producer
	snapshotProducer *Producer
	deltaProducer    *Producer
	lastSnapshot     map[string]domain.OrderBookSnapshot
	store            SnapshotStore
}

func NewConsumer(
	brokers []string,
	groupID string,
	topic string,
	matcher *application.Matcher,
	tradeProducer *Producer,
	snapshotProducer *Producer,
	deltaProducer *Producer,
	store SnapshotStore,
) *Consumer {

	return &Consumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers:        brokers,
			GroupID:        groupID,
			Topic:          topic,
			MinBytes:       10e3,
			MaxBytes:       10e6,
			CommitInterval: 0, // manual commit
		}),
		matcher:          matcher,
		tradeProducer:    tradeProducer,
		snapshotProducer: snapshotProducer,
		deltaProducer:    deltaProducer,
		lastSnapshot:     make(map[string]domain.OrderBookSnapshot),
		store:            store,
	}
}

func (c *Consumer) Start(ctx context.Context) {
	for {
		m, err := c.reader.FetchMessage(ctx)
		if err != nil {
			log.Println("error fetching message:", err)
			continue
		}

		var order domain.Order
		if err := contracts.ValidateTopicPayload("orders", m.Value); err != nil {
			log.Println("order schema validation failed:", err)
			continue
		}
		if err := json.Unmarshal(m.Value, &order); err != nil {
			log.Println("invalid order payload:", err)
			continue
		}
		traceID := readHeader(m.Headers, "trace-id")

		trades := c.matcher.Process(order)

		for _, trade := range trades {
			if err := c.tradeProducer.PublishTradeWithTrace(ctx, trade, traceID); err != nil {
				log.Println("failed to publish trade:", err)
				continue
			}
		}

		snapshot := c.matcher.GetSnapshot(order.Symbol, 10)
		if err := c.snapshotProducer.PublishSnapshotWithTrace(ctx, snapshot, traceID); err != nil {
			log.Println("failed to publish snapshot:", err)
		}
		if c.store != nil {
			if err := c.store.Save(order.Symbol, snapshot); err != nil {
				log.Println("failed to persist snapshot:", err)
			}
		}

		prev := c.lastSnapshot[order.Symbol]
		delta := buildDelta(snapshot, prev)
		c.lastSnapshot[order.Symbol] = snapshot
		if err := c.deltaProducer.PublishDeltaWithTrace(ctx, delta, traceID); err != nil {
			log.Println("failed to publish delta:", err)
		}

		// Commit AFTER successful processing
		if err := c.reader.CommitMessages(ctx, m); err != nil {
			log.Println("commit failed:", err)
		}
	}
}

func readHeader(headers []kafka.Header, key string) string {
	for _, h := range headers {
		if h.Key == key {
			return string(h.Value)
		}
	}
	return ""
}

func buildDelta(curr, prev domain.OrderBookSnapshot) domain.OrderBookDelta {
	return domain.OrderBookDelta{
		Symbol:    curr.Symbol,
		Bids:      diffLevels(curr.Bids, prev.Bids),
		Asks:      diffLevels(curr.Asks, prev.Asks),
		Spread:    curr.Spread,
		Timestamp: time.Now().Unix(),
	}
}

func diffLevels(curr, prev []domain.Level) []domain.Level {
	currMap := make(map[float64]int, len(curr))
	prevMap := make(map[float64]int, len(prev))

	for _, l := range curr {
		currMap[l.Price] = l.Quantity
	}
	for _, l := range prev {
		prevMap[l.Price] = l.Quantity
	}

	changes := make([]domain.Level, 0)

	for price, qty := range currMap {
		if prevQty, ok := prevMap[price]; !ok || prevQty != qty {
			changes = append(changes, domain.Level{
				Price:    price,
				Quantity: qty,
			})
		}
	}

	for price := range prevMap {
		if _, ok := currMap[price]; !ok {
			changes = append(changes, domain.Level{
				Price:    price,
				Quantity: 0,
			})
		}
	}

	return changes
}
