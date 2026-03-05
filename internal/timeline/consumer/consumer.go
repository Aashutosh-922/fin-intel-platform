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

	events "github.com/Aashutosh-922/fin-intel-platform/internal/timeline/application/events"
	"github.com/twmb/franz-go/pkg/kgo"
)

type RiskDecision struct {
	TransactionID string `json:"transaction_id"`
	RiskScore     int    `json:"risk_score"`
	Decision      string `json:"decision"`
}

type AIInsight struct {
	TransactionID string  `json:"transaction_id"`
	Verdict       string  `json:"verdict"`
	Confidence    float64 `json:"confidence"`
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
			switch record.Topic {
			case "risk-decisions":
				c.handleRiskRecord(ctx, record)
			case "ai-insights":
				c.handleAIRecord(ctx, record)
			default:
				log.Printf("ignoring unsupported topic %s", record.Topic)
			}
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

func (c *DecisionConsumer) handleRiskRecord(ctx context.Context, record *kgo.Record) {
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

func (c *DecisionConsumer) handleAIRecord(ctx context.Context, record *kgo.Record) {
	var insight AIInsight

	if err := json.Unmarshal(record.Value, &insight); err != nil {
		log.Println("invalid ai insight event")
		return
	}

	if insight.TransactionID == "" {
		log.Println("ai insight missing transaction_id")
		return
	}

	err := c.service.Record(ctx, events.Event{
		TransactionID: insight.TransactionID,
		Type:          "AI_ANALYSIS",
		Payload:       string(record.Value),
	})
	if err != nil {
		log.Println("failed writing ai insight timeline:", err)
	}
}
