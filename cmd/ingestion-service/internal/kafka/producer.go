// package kafka

// import (
// 	"context"
// 	"encoding/json"
// 	"log"
// 	"time"

// 	"github.com/segmentio/kafka-go"
// )

// type Producer struct {
// 	writer *kafka.Writer
// 	topic  string
// }

// func NewProducer(brokers []string, topic string) *Producer {
// 	return &Producer{
// 		topic: topic,
// 		writer: &kafka.Writer{
// 			Addr:         kafka.TCP(brokers...),
// 			Topic:        topic,
// 			Balancer:     &kafka.LeastBytes{},
// 			RequiredAcks: kafka.RequireAll,
// 			Async:        false, // 🔴 IMPORTANT
// 		},
// 	}
// }

// // func (p *Producer) Publish(ctx context.Context, payload any) error {
// // 	bytes, err := json.Marshal(payload)
// // 	if err != nil {
// // 		return err
// // 	}

// // 	msg := kafka.Message{
// // 		Key:   []byte(time.Now().String()),
// // 		Value: bytes,
// // 	}

// // 	err = p.writer.WriteMessages(ctx, msg)
// // 	if err != nil {
// // 		log.Println("❌ kafka publish failed:", err)
// // 		return err
// // 	}

// // 	log.Println("✅ published to kafka:", p.topic)
// // 	return nil
// // }

// func (p *Producer) Publish(ctx context.Context, payload any) error {
// 	bytes, err := json.Marshal(payload)
// 	if err != nil {
// 		return err
// 	}

// 	log.Println("➡️ kafka publish attempt", "topic", p.topic, "payload", string(bytes))

// 	msg := kafka.Message{
// 		Key:   []byte(time.Now().String()),
// 		Value: bytes,
// 	}

// 	err = p.writer.WriteMessages(ctx, msg)
// 	if err != nil {
// 		log.Println("❌ kafka publish failed:", err)
// 		return err
// 	}

// 	log.Println("✅ kafka publish success", "topic", p.topic)
// 	return nil
// }

package kafka

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
)

type Producer struct {
	mainWriter *kafka.Writer
	dlqWriter  *kafka.Writer
	mainTopic  string
	dlqTopic   string
}

func NewProducer(brokers []string, mainTopic, dlqTopic string) *Producer {
	return &Producer{
		mainTopic: mainTopic,
		dlqTopic:  dlqTopic,

		mainWriter: &kafka.Writer{
			Addr:         kafka.TCP(brokers...),
			Topic:        mainTopic,
			Balancer:     &kafka.LeastBytes{},
			RequiredAcks: kafka.RequireAll,
			Async:        false,
		},

		dlqWriter: &kafka.Writer{
			Addr:         kafka.TCP(brokers...),
			Topic:        dlqTopic,
			Balancer:     &kafka.LeastBytes{},
			RequiredAcks: kafka.RequireAll,
			Async:        false,
		},
	}
}

func (p *Producer) Publish(ctx context.Context, payload any) error {
	bytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	msg := kafka.Message{
		Key:   []byte(time.Now().UTC().String()),
		Value: bytes,
		Time:  time.Now().UTC(),
	}

	const maxRetries = 3
	var lastErr error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		err = p.mainWriter.WriteMessages(ctx, msg)
		if err == nil {
			log.Println("✅ kafka publish success:", p.mainTopic)
			return nil
		}

		lastErr = err
		log.Printf("⚠️ kafka publish failed (attempt %d/%d): %v\n", attempt, maxRetries, err)

		time.Sleep(time.Duration(attempt) * time.Second) // simple backoff
	}

	// =========================
	// ❌ SEND TO DLQ
	// =========================

	log.Println("🚨 sending message to DLQ:", p.dlqTopic)

	dlqErr := p.dlqWriter.WriteMessages(ctx, msg)
	if dlqErr != nil {
		log.Println("❌ DLQ publish ALSO failed:", dlqErr)
		return dlqErr
	}

	log.Println("🧟 message sent to DLQ after retries")
	return lastErr
}