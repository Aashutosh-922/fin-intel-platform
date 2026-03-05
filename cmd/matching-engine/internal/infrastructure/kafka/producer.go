package kafka

import (
	"context"
	"encoding/json"

	"github.com/Aashutosh-922/fin-intel-platform/cmd/matching-engine/internal/domain"
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
			Async:        false,
		},
	}
}

func (p *Producer) Publish(ctx context.Context, trade domain.Trade) error {
	return p.PublishTrade(ctx, trade)
}

func (p *Producer) PublishTrade(ctx context.Context, trade domain.Trade) error {
	return p.PublishTradeWithTrace(ctx, trade, "")
}

func (p *Producer) PublishTradeWithTrace(ctx context.Context, trade domain.Trade, traceID string) error {
	payload, err := json.Marshal(trade)
	if err != nil {
		return err
	}

	msg := kafka.Message{
		Key:   []byte(trade.Symbol), // partition by symbol
		Value: payload,
	}
	if traceID != "" {
		msg.Headers = append(msg.Headers, kafka.Header{
			Key:   "trace-id",
			Value: []byte(traceID),
		})
	}

	return p.writer.WriteMessages(ctx, msg)
}

func (p *Producer) PublishSnapshot(ctx context.Context, snapshot domain.OrderBookSnapshot) error {
	return p.PublishSnapshotWithTrace(ctx, snapshot, "")
}

func (p *Producer) PublishSnapshotWithTrace(ctx context.Context, snapshot domain.OrderBookSnapshot, traceID string) error {
	payload, err := json.Marshal(snapshot)
	if err != nil {
		return err
	}

	msg := kafka.Message{
		Key:   []byte(snapshot.Symbol),
		Value: payload,
	}
	if traceID != "" {
		msg.Headers = append(msg.Headers, kafka.Header{
			Key:   "trace-id",
			Value: []byte(traceID),
		})
	}

	return p.writer.WriteMessages(ctx, msg)
}

func (p *Producer) PublishDelta(ctx context.Context, delta domain.OrderBookDelta) error {
	return p.PublishDeltaWithTrace(ctx, delta, "")
}

func (p *Producer) PublishDeltaWithTrace(ctx context.Context, delta domain.OrderBookDelta, traceID string) error {
	payload, err := json.Marshal(delta)
	if err != nil {
		return err
	}

	msg := kafka.Message{
		Key:   []byte(delta.Symbol),
		Value: payload,
	}
	if traceID != "" {
		msg.Headers = append(msg.Headers, kafka.Header{
			Key:   "trace-id",
			Value: []byte(traceID),
		})
	}

	return p.writer.WriteMessages(ctx, msg)
}

func (p *Producer) Close() error {
	return p.writer.Close()
}

func (p *Producer) PublishOrderEvent(ctx context.Context, event domain.OrderEvent) error {

	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}

	msg := kafka.Message{
		Key:   []byte(event.OrderID),
		Value: payload,
	}

	return p.writer.WriteMessages(ctx, msg)
}
