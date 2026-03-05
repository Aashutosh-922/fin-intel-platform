package kafka

import (
	"context"
	"log"

	"github.com/segmentio/kafka-go"
)

type TopicConsumer struct {
	reader   *kafka.Reader
	onRecord func(kafka.Message)
}

func NewTopicConsumer(brokers []string, groupID, topic string, onRecord func(kafka.Message)) *TopicConsumer {
	return &TopicConsumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers: brokers,
			GroupID: groupID,
			Topic:   topic,
		}),
		onRecord: onRecord,
	}
}

func (c *TopicConsumer) Start(ctx context.Context) {
	for {
		msg, err := c.reader.FetchMessage(ctx)
		if err != nil {
			log.Printf("consumer fetch error (%s): %v", c.reader.Config().Topic, err)
			continue
		}

		c.onRecord(msg)

		if err := c.reader.CommitMessages(ctx, msg); err != nil {
			log.Printf("consumer commit error (%s): %v", c.reader.Config().Topic, err)
		}
	}
}
