package kafka

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/Aashutosh-922/fin-intel-platform/cmd/order-service/internal/domain"
	otrace "github.com/Aashutosh-922/fin-intel-platform/cmd/order-service/internal/trace"
	"github.com/segmentio/kafka-go"
)

type Producer struct {
	orderWriter  *kafka.Writer
	cancelWriter *kafka.Writer
}

func NewProducer(brokersCSV string, orderTopic string) *Producer {
	brokers := strings.Split(brokersCSV, ",")
	return &Producer{
		orderWriter:  newWriter(brokers, orderTopic),
		cancelWriter: newWriter(brokers, "order-cancel"),
	}
}

func newWriter(brokers []string, topic string) *kafka.Writer {
	return &kafka.Writer{
		Addr:         kafka.TCP(brokers...),
		Topic:        topic,
		RequiredAcks: kafka.RequireAll,
		Async:        false,
	}
}

func (p *Producer) Publish(ctx context.Context, event domain.OrderCreatedEvent) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}

	msg := kafka.Message{
		Key:   []byte(event.Symbol),
		Value: payload,
	}
	if traceID := otrace.TraceID(ctx); traceID != "" {
		msg.Headers = append(msg.Headers, kafka.Header{
			Key:   "trace-id",
			Value: []byte(traceID),
		})
	}

	return p.orderWriter.WriteMessages(ctx, msg)
}

func (p *Producer) PublishCancel(ctx context.Context, cancel domain.OrderCancel) error {
	payload, err := json.Marshal(cancel)
	if err != nil {
		return err
	}

	msg := kafka.Message{
		Key:   []byte(cancel.Symbol),
		Value: payload,
	}
	if traceID := otrace.TraceID(ctx); traceID != "" {
		msg.Headers = append(msg.Headers, kafka.Header{
			Key:   "trace-id",
			Value: []byte(traceID),
		})
	}

	return p.cancelWriter.WriteMessages(ctx, msg)
}

func (p *Producer) Close() error {
	var err error
	if e := p.orderWriter.Close(); e != nil {
		err = e
	}
	if e := p.cancelWriter.Close(); e != nil && err == nil {
		err = e
	}
	return err
}
