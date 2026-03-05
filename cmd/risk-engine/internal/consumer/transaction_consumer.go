package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"

	"github.com/Aashutosh-922/fin-intel-platform/cmd/risk-engine/internal/domain/risk"
	"github.com/Aashutosh-922/fin-intel-platform/cmd/risk-engine/internal/producer"

	"github.com/Aashutosh-922/fin-intel-platform/cmd/risk-engine/internal/rules"
	"github.com/Aashutosh-922/fin-intel-platform/internal/events"
)

type Consumer struct {
	client   *kgo.Client
	producer *producer.Producer
	scorer   *risk.Scorer
}

func New(
	client *kgo.Client,
	producer *producer.Producer,
	_ any, // logger already handled at app level
) *Consumer {
	return &Consumer{
		client:   client,
		producer: producer,
		scorer: risk.NewScorer([]risk.Rule{
			rules.AmountSpikeRule{},
			rules.GeoMismatchRule{},
			rules.VelocityRule{},
		}),
	}
}

func (c *Consumer) Start(ctx context.Context) {
	for {
		fetches := c.client.PollFetches(ctx)
		if fetches.IsClientClosed() {
			return
		}

		fetches.EachPartition(func(p kgo.FetchTopicPartition) {
			for _, record := range p.Records {

				err := c.handleRecord(ctx, record)

				if err == nil {
					// ✅ success → commit offset
					c.client.CommitRecords(ctx, record)
					continue
				}

				// ❌ failure path
				retries := getRetryCount(record)

				if retries < 3 {
					log.Printf("retrying record (attempt %d)", retries+1)
					c.retry(ctx, record, retries+1)
				} else {
					log.Printf("sending to DLQ after %d retries: %v", retries, err)
					c.sendToDLQ(ctx, record, err)
					c.client.CommitRecords(ctx, record) // commit poison message
				}
			}
		})
	}
}

// func (c *Consumer) handleRecord(ctx context.Context, record *kgo.Record) error {
// 	var event events.TransactionEvent

// 	if err := json.Unmarshal(record.Value, &event); err != nil {
// 		return err
// 	}

// 	log.Printf("received transaction: %+v", event)

// 	riskCtx := buildContext(event)

// 	score, factors := c.scorer.Score(riskCtx)

// 	decision := risk.Decide(score)

// 	log.Printf("risk evaluated: user=%s score=%d decision=%s factors=%v",
// 		event.UserID,
// 		score,
// 		decision,
// 		factors,
// 	)

// 	return emitDecision(ctx, c.producer, event.EventID, score, decision)
// }
// func (c *Consumer) handleRecord(ctx context.Context, record *kgo.Record) {
// 	retries := getRetryCount(record)

// 	var evt events.TransactionEvent
// 	if err := json.Unmarshal(record.Value, &evt); err != nil {
// 		log.Println("invalid json → DLQ")
// 		c.sendToDLQ(ctx, record, err)
// 		return
// 	}

// 	log.Printf("received transaction: %+v", evt)

// 	riskCtx := buildContext(evt)

// 	score, factors := c.scorer.Score(riskCtx)
// 	decision := risk.Decide(score)

// 	log.Printf("risk evaluated: user=%s score=%d decision=%s factors=%v",
// 		evt.UserID, score, decision, factors)

// 	err := emitDecision(ctx, c.producer, evt.EventID, score, decision)
// 	if err != nil {
// 		if retries < 3 {
// 			log.Println("temporary failure → retry", retries+1)
// 			c.retry(ctx, record, retries+1)
// 		} else {
// 			log.Println("max retries reached → DLQ")
// 			c.sendToDLQ(ctx, record, err)
// 		}
// 	}
// }
// func (c *Consumer) handleRecord(ctx context.Context, record *kgo.Record) error {
// 	retries := getRetryCount(record)

// 	var event events.TransactionEvent
// 	if err := json.Unmarshal(record.Value, &event); err != nil {
// 		log.Println("invalid json → DLQ")
// 		c.sendToDLQ(ctx, record, err)
// 		return nil
// 	}

// 	riskCtx := buildContext(event)

// 	score, _ := c.scorer.Score(riskCtx)

//     decision := risk.Decide(score)

//     log.Printf("risk evaluated: user=%s score=%d decision=%s",
//     event.UserID, score, decision)

// 	err := emitDecision(ctx, c.producer, event.EventID, score, decision)
// 	if err != nil {
// 		if retries < 3 {
// 			log.Println("temporary failure → retry", retries+1)
// 			c.retry(ctx, record, retries+1)
// 		} else {
// 			log.Println("max retries reached → DLQ")
// 			c.sendToDLQ(ctx, record, err)
// 		}
// 		return nil
// 	}

//		return nil
//	}
func (c *Consumer) handleRecord(ctx context.Context, record *kgo.Record) error {
	retries := getRetryCount(record)

	var event events.TransactionEvent

	// 1️⃣ SCHEMA ERROR → DLQ (no retry)
	if err := json.Unmarshal(record.Value, &event); err != nil {
		log.Println("invalid json → DLQ")
		c.sendToDLQ(ctx, record, err)
		return nil
	}

	log.Printf("received transaction: %+v", event)

	// 2️⃣ PROCESSING
	riskCtx := buildContext(event)
	score, factors := c.scorer.Score(riskCtx)
	decision := risk.Decide(score)

	log.Printf("risk evaluated: user=%s score=%d decision=%s factors=%v",
		event.UserID, score, decision, factors)

	// 3️⃣ PRODUCE RESULT
	err := emitDecision(ctx, c.producer, event.EventID, score, decision)
	if err != nil {
		// retryable failure
		if retries < 3 {
			log.Printf("retrying record (attempt %d)", retries+1)
			c.retry(ctx, record, retries+1)
		} else {
			log.Println("sending to DLQ after max retries")
			c.sendToDLQ(ctx, record, err)
		}
		return nil
	}

	return nil
}

// -------- Helper functions --------
func buildContext(event events.TransactionEvent) risk.Context {
	return risk.Context{
		TransactionID: event.EventID,
		UserID:        event.UserID,
		Amount:        event.Amount,
	}
}

func emitDecision(
	ctx context.Context,
	p *producer.Producer,
	txID string,
	score int,
	decision string,
) error {
	return p.Publish(ctx, producer.RiskEvent{
		EventID:       txID + "-" + decision,
		TransactionID: txID,
		RiskScore:     score,
		Decision:      decision,
		Flagged:       decision != "APPROVED",
		CreatedAt:     time.Now().Unix(),
	})
}

func getRetryCount(record *kgo.Record) int {
	for _, h := range record.Headers {
		if string(h.Key) == "retry-count" {
			var n int
			fmt.Sscanf(string(h.Value), "%d", &n)
			return n
		}
	}
	return 0
}

func (c *Consumer) retry(ctx context.Context, record *kgo.Record, count int) {
	record.Topic = "transactions-retry"

	record.Headers = append(record.Headers, kgo.RecordHeader{
		Key:   "retry-count",
		Value: []byte(fmt.Sprintf("%d", count)),
	})

	c.client.Produce(ctx, record, nil)
}

func (c *Consumer) sendToDLQ(ctx context.Context, record *kgo.Record, reason error) {
	dlqRecord := &kgo.Record{
		Topic: "transactions-dlq",
		Value: record.Value,
		Headers: append(record.Headers,
			kgo.RecordHeader{Key: "error", Value: []byte(reason.Error())},
		),
	}

	c.client.Produce(ctx, dlqRecord, nil)
}
