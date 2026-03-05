package gemini

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
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
	APIKey     string
	Model      string
	BaseURL    string
	TimeoutMs  int
	MaxRetries int
}

func NewClient(cfg Config) *Client {
	timeout := time.Duration(cfg.TimeoutMs) * time.Millisecond
	if timeout <= 0 {
		timeout = 4 * time.Second
	}
	baseURL := normalizeBaseURL(cfg.BaseURL)
	model := strings.TrimSpace(cfg.Model)
	if model == "" {
		model = "gemini-1.5-flash"
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

func normalizeBaseURL(input string) string {
	baseURL := strings.TrimRight(strings.TrimSpace(input), "/")
	if baseURL == "" {
		return "https://generativelanguage.googleapis.com/v1beta"
	}

	parsed, err := url.Parse(baseURL)
	if err != nil {
		return baseURL
	}

	path := strings.TrimRight(parsed.Path, "/")
	if path == "" {
		parsed.Path = "/v1beta"
		return strings.TrimRight(parsed.String(), "/")
	}
	return baseURL
}

func (c *Client) Enabled() bool {
	return c.apiKey != ""
}

func (c *Client) Analyze(ctx context.Context, event domain.RiskDecisionEvent) (domain.AIAnalysis, error) {
	if !c.Enabled() {
		return domain.AIAnalysis{}, errors.New("gemini client not configured")
	}

	prompt := fmt.Sprintf(
		"Analyze risk decision event and enrich rationale.\ntransaction_id=%s\nevent_id=%s\nrisk_score=%s\nflagged=%t\ndecision_created_at=%d",
		event.TransactionID,
		event.EventID,
		strconv.FormatFloat(event.RiskScore, 'f', -1, 64),
		event.Flagged,
		event.CreatedAt,
	)

	reqBody := map[string]any{
		"system_instruction": map[string]any{
			"parts": []map[string]string{{
				"text": "You are a financial risk analyst. Return strict JSON with keys: verdict, confidence, reasoning. confidence must be [0,1], reasoning must be an array of short strings.",
			}},
		},
		"contents": []map[string]any{{
			"role": "user",
			"parts": []map[string]string{{
				"text": prompt,
			}},
		}},
		"generationConfig": map[string]any{
			"temperature":      0.1,
			"responseMimeType": "application/json",
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
			time.Sleep(time.Duration(200*(attempt+1)) * time.Millisecond)
		}
	}
	return domain.AIAnalysis{}, lastErr
}

func (c *Client) call(ctx context.Context, reqBody map[string]any, event domain.RiskDecisionEvent) (domain.AIAnalysis, error) {
	body, err := json.Marshal(reqBody)
	if err != nil {
		return domain.AIAnalysis{}, err
	}

	url := fmt.Sprintf("%s/models/%s:generateContent?key=%s", c.baseURL, c.model, c.apiKey)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return domain.AIAnalysis{}, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return domain.AIAnalysis{}, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return domain.AIAnalysis{}, fmt.Errorf("gemini status=%d body=%s", resp.StatusCode, string(respBody))
	}

	var parsed geminiResponse
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return domain.AIAnalysis{}, err
	}
	if len(parsed.Candidates) == 0 || len(parsed.Candidates[0].Content.Parts) == 0 {
		return domain.AIAnalysis{}, errors.New("gemini response has no content")
	}

	content := ""
	for _, part := range parsed.Candidates[0].Content.Parts {
		if strings.TrimSpace(part.Text) != "" {
			content = part.Text
			break
		}
	}
	if content == "" {
		return domain.AIAnalysis{}, errors.New("gemini response has empty text content")
	}

	content = extractJSONObject(content)
	var output llmOutput
	if err := json.Unmarshal([]byte(content), &output); err != nil {
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
		EventID:       "llm-" + strconv.FormatInt(time.Now().UnixNano(), 10),
		TransactionID: event.TransactionID,
		Verdict:       output.Verdict,
		Confidence:    output.Confidence,
		Reasoning:     output.Reasoning,
		CreatedAt:     time.Now().Unix(),
	}, nil
}

type geminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
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

func extractJSONObject(s string) string {
	trimmed := strings.TrimSpace(s)
	if trimmed == "" {
		return trimmed
	}
	if strings.HasPrefix(trimmed, "{") && strings.HasSuffix(trimmed, "}") {
		return trimmed
	}
	start := strings.Index(trimmed, "{")
	end := strings.LastIndex(trimmed, "}")
	if start >= 0 && end > start {
		return trimmed[start : end+1]
	}
	return trimmed
}
