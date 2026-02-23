// package consumer

// import (
// 	"context"
// 	"encoding/json"
// 	"log"

// 	"github.com/twmb/franz-go/pkg/kgo"

// 	events "github.com/Aashutosh-922/fin-intel-platform/internal/timeline/application/events"
// )

// type RiskDecision struct {
// 	TransactionID string `json:"transaction_id"`
// 	RiskScore     int    `json:"risk_score"`
// 	Decision      string `json:"decision"`
// }

// type DecisionConsumer struct {
// 	client  *kgo.Client
// 	service *events.Service
// }

// func NewDecisionConsumer(client *kgo.Client, service *events.Service) *DecisionConsumer {
// 	return &DecisionConsumer{client: client, service: service}
// }

// func (c *DecisionConsumer) Start(ctx context.Context) {
// 	for {
// 		fetches := c.client.PollFetches(ctx)
// 		if fetches.IsClientClosed() {
// 			return
// 		}

// 		fetches.EachRecord(func(record *kgo.Record) {
// 			c.handle(ctx, record)
// 		})
// 	}
// }

// func (c *DecisionConsumer) handle(ctx context.Context, record *kgo.Record) {
// 	var decision RiskDecision

// 	if err := json.Unmarshal(record.Value, &decision); err != nil {
// 		log.Println("invalid decision event")
// 		return
// 	}

// 	log.Printf("timeline event: %+v\n", decision)

// 	err := c.service.Record(ctx, app.Event{
// 		ID:            decision.TransactionID + "-" + decision.Decision,
// 		TransactionID: decision.TransactionID,
// 		Type:          decision.Decision,
// 		Payload:       string(record.Value),
// 		CreatedAt:     time.Now().UTC(),
// 	})	

// 	if err != nil {
// 		log.Println("failed writing timeline:", err)
// 	}
// }

package consumer

import (
	"context"
	"encoding/json"
	"log"

	"github.com/twmb/franz-go/pkg/kgo"
	events "github.com/Aashutosh-922/fin-intel-platform/internal/timeline/application/events"
)

type RiskDecision struct {
	TransactionID string `json:"transaction_id"`
	RiskScore     int    `json:"risk_score"`
	Decision      string `json:"decision"`
}

type DecisionConsumer struct {
	client  *kgo.Client
	service *events.Service
}

func NewDecisionConsumer(client *kgo.Client, service *events.Service) *DecisionConsumer {
	return &DecisionConsumer{client: client, service: service}
}

func (c *DecisionConsumer) Start(ctx context.Context) {
	for {
		fetches := c.client.PollFetches(ctx)
		if fetches.IsClientClosed() {
			return
		}

		fetches.EachRecord(func(record *kgo.Record) {
			c.handle(ctx, record)
		})
	}
}

// func (c *DecisionConsumer) handle(ctx context.Context, record *kgo.Record) {
// 	var decision RiskDecision

// 	if err := json.Unmarshal(record.Value, &decision); err != nil {
// 		log.Println("invalid decision event")
// 		return
// 	}

// 	log.Printf("timeline event: %+v\n", decision)

// 	// err := c.service.Record(ctx, events.Event{
// 	// 	ID:            decision.TransactionID + "-" + decision.Decision,
// 	// 	TransactionID: decision.TransactionID,
// 	// 	Type:          decision.Decision,
// 	// 	Payload:       string(record.Value),
// 	// 	CreatedAt:     time.Now().UTC(),
// 	// })
// 	err := c.service.Record(ctx, events.Event{
// 		EventID:       decision.TransactionID + "-" + decision.Decision,
// 		TransactionID: decision.TransactionID,
// 		Type:          decision.Decision,
// 		Payload:       string(record.Value),
// 		CreatedAt:     time.Now().UTC(),
// 	})

// 	if err != nil {
// 		log.Println("failed writing timeline:", err)
// 	}
// }

// func (c *DecisionConsumer) handle(ctx context.Context, record *kgo.Record) {
// 	var decision RiskDecision

// 	if err := json.Unmarshal(record.Value, &decision); err != nil {
// 		log.Println("invalid decision event")
// 		return
// 	}

// 	log.Printf("timeline event: %+v\n", decision)

// 	err := c.service.Record(ctx, events.Event{
// 		TransactionID: decision.TransactionID,
// 		Type:          decision.Decision,
// 		Payload:       string(record.Value),
// 	})

// 	if err != nil {
// 		log.Println("failed writing timeline:", err)
// 	}
// }

func (c *DecisionConsumer) handle(ctx context.Context, record *kgo.Record) {
	var decision RiskDecision

	if err := json.Unmarshal(record.Value, &decision); err != nil {
		log.Println("invalid decision event")
		return
	}

	log.Printf("timeline event: %+v\n", decision)

	err := c.service.Record(ctx, events.Event{
		TransactionID: decision.TransactionID,
		Type:          decision.Decision,
		Payload:       string(record.Value),
	})

	if err != nil {
		log.Println("failed writing timeline:", err)
	}
}