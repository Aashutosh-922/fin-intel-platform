package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Aashutosh-922/fin-intel-platform/cmd/ai-service/internal/domain"
)

type Client struct {
	apiKey     string
	model      string
	baseURL    string
	httpClient *http.Client
	maxRetries int
}

type Config struct {
	APIKey      string
	Model       string
	BaseURL     string
	TimeoutMs   int
	MaxRetries  int
	Temperature float64
}

func NewClient(cfg Config) *Client {
	timeout := time.Duration(cfg.TimeoutMs) * time.Millisecond
	if timeout <= 0 {
		timeout = 4 * time.Second
	}
	baseURL := strings.TrimRight(strings.TrimSpace(cfg.BaseURL), "/")
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}
	model := strings.TrimSpace(cfg.Model)
	if model == "" {
		model = "gpt-4.1-mini"
	}
	retries := cfg.MaxRetries
	if retries < 0 {
		retries = 0
	}
	if retries > 4 {
		retries = 4
	}
	return &Client{
		apiKey:  strings.TrimSpace(cfg.APIKey),
		model:   model,
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: timeout,
		},
		maxRetries: retries,
	}
}

func (c *Client) Enabled() bool {
	return c.apiKey != ""
}

func (c *Client) Analyze(ctx context.Context, event domain.RiskDecisionEvent) (domain.AIAnalysis, error) {
	if !c.Enabled() {
		return domain.AIAnalysis{}, errors.New("openai client not configured")
	}

	reqBody := map[string]any{
		"model":       c.model,
		"temperature": 0.1,
		"response_format": map[string]any{
			"type": "json_object",
		},
		"messages": []map[string]string{
			{
				"role": "system",
				"content": "You are a financial risk analyst. Return strict JSON with keys: verdict, confidence, reasoning." +
					" confidence must be [0,1], reasoning must be an array of short strings.",
			},
			{
				"role": "user",
				"content": fmt.Sprintf(
					"Analyze risk decision event and enrich rationale.\ntransaction_id=%s\nevent_id=%s\nrisk_score=%s\nflagged=%t\ndecision_created_at=%d",
					event.TransactionID,
					event.EventID,
					strconv.FormatFloat(event.RiskScore, 'f', -1, 64),
					event.Flagged,
					event.CreatedAt,
				),
			},
		},
	}

	var lastErr error
	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		analysis, err := c.call(ctx, reqBody, event)
		if err == nil {
			return analysis, nil
		}
		lastErr = err
		if attempt < c.maxRetries {
			sleep := time.Duration(200*(attempt+1)) * time.Millisecond
			time.Sleep(sleep)
		}
	}
	return domain.AIAnalysis{}, lastErr
}

func (c *Client) call(ctx context.Context, reqBody map[string]any, event domain.RiskDecisionEvent) (domain.AIAnalysis, error) {
	body, err := json.Marshal(reqBody)
	if err != nil {
		return domain.AIAnalysis{}, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return domain.AIAnalysis{}, err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return domain.AIAnalysis{}, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return domain.AIAnalysis{}, fmt.Errorf("openai status=%d body=%s", resp.StatusCode, string(respBody))
	}

	var parsed chatResponse
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return domain.AIAnalysis{}, err
	}
	if len(parsed.Choices) == 0 {
		return domain.AIAnalysis{}, errors.New("openai response has no choices")
	}

	var output llmOutput
	if err := json.Unmarshal([]byte(parsed.Choices[0].Message.Content), &output); err != nil {
		return domain.AIAnalysis{}, err
	}
	if output.Verdict == "" {
		output.Verdict = "Low Risk"
	}
	if len(output.Reasoning) == 0 {
		output.Reasoning = []string{"Model returned no reasoning"}
	}
	output.Confidence = clamp01(output.Confidence)

	return domain.AIAnalysis{
		EventID:       generateID(),
		TransactionID: event.TransactionID,
		Verdict:       output.Verdict,
		Confidence:    output.Confidence,
		Reasoning:     output.Reasoning,
		CreatedAt:     now(),
	}, nil
}

type chatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

type llmOutput struct {
	Verdict    string   `json:"verdict"`
	Confidence float64  `json:"confidence"`
	Reasoning  []string `json:"reasoning"`
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

func now() int64 {
	return time.Now().Unix()
}

func generateID() string {
	return "llm-" + strconv.FormatInt(time.Now().UnixNano(), 10)
}
