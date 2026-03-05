package clients

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/Aashutosh-922/fin-intel-platform/cmd/api-gateway/internal/handlers"
)

type AIClient struct {
	eventDB *sql.DB
}

func NewAIClient(eventDB *sql.DB) *AIClient {
	return &AIClient{eventDB: eventDB}
}

func (c *AIClient) Query(
	ctx context.Context,
	q handlers.AIQuery,
) (handlers.AIResponse, error) {
	type aiPayload struct {
		Verdict    string   `json:"verdict"`
		Confidence float64  `json:"confidence"`
		Reasoning  []string `json:"reasoning"`
	}

	var raw string
	err := c.eventDB.QueryRowContext(
		ctx,
		`SELECT payload
		 FROM transaction_events
		 WHERE transaction_id = $1
		   AND event_type = 'AI_ANALYSIS'
		 ORDER BY created_at DESC
		 LIMIT 1`,
		q.TransactionID,
	).Scan(&raw)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return handlers.AIResponse{}, sql.ErrNoRows
		}
		return handlers.AIResponse{}, err
	}

	var payload aiPayload
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		return handlers.AIResponse{}, err
	}

	return handlers.AIResponse{
		Text: fmt.Sprintf("verdict=%s confidence=%.3f reasoning=%v", payload.Verdict, payload.Confidence, payload.Reasoning),
	}, nil
}
