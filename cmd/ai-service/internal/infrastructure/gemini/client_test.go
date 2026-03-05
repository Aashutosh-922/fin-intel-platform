package gemini

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/Aashutosh-922/fin-intel-platform/cmd/ai-service/internal/domain"
)

func TestAnalyzeSuccessWithBaseURLNormalization(t *testing.T) {
	var hitCount int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&hitCount, 1)

		if r.Method != http.MethodPost {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/v1beta/models/gemini-test:generateContent" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("key") != "test-key" {
			t.Fatalf("missing api key query param")
		}

		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("failed to decode request payload: %v", err)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"candidates":[{"content":{"parts":[{"text":"Output:\n{\"verdict\":\"High Risk\",\"confidence\":1.4,\"reasoning\":[\"rule + llm\"]}"}]}}]}`))
	}))
	defer srv.Close()

	client := NewClient(Config{
		APIKey:  "test-key",
		Model:   "gemini-test",
		BaseURL: srv.URL, // root URL should be normalized to /v1beta
	})

	event := domain.RiskDecisionEvent{
		EventID:       "evt-1",
		TransactionID: "tx-1",
		RiskScore:     95,
		Flagged:       true,
		CreatedAt:     1,
	}

	got, err := client.Analyze(context.Background(), event)
	if err != nil {
		t.Fatalf("Analyze returned error: %v", err)
	}

	if got.TransactionID != event.TransactionID {
		t.Fatalf("unexpected transaction id: %s", got.TransactionID)
	}
	if got.Verdict != "High Risk" {
		t.Fatalf("unexpected verdict: %s", got.Verdict)
	}
	if got.Confidence != 1 {
		t.Fatalf("expected confidence clamp to 1, got %v", got.Confidence)
	}
	if len(got.Reasoning) != 1 || got.Reasoning[0] != "rule + llm" {
		t.Fatalf("unexpected reasoning: %#v", got.Reasoning)
	}
	if atomic.LoadInt32(&hitCount) != 1 {
		t.Fatalf("expected one request, got %d", hitCount)
	}
}

func TestAnalyzeRetriesOnFailure(t *testing.T) {
	var hitCount int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cur := atomic.AddInt32(&hitCount, 1)
		if cur == 1 {
			http.Error(w, "try again", http.StatusBadGateway)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"candidates":[{"content":{"parts":[{"text":"{\"verdict\":\"Low Risk\",\"confidence\":0.7,\"reasoning\":[\"ok\"]}"}]}}]}`))
	}))
	defer srv.Close()

	client := NewClient(Config{
		APIKey:     "test-key",
		Model:      "gemini-test",
		BaseURL:    srv.URL,
		MaxRetries: 1,
	})

	_, err := client.Analyze(context.Background(), domain.RiskDecisionEvent{TransactionID: "tx-1"})
	if err != nil {
		t.Fatalf("Analyze returned error: %v", err)
	}
	if atomic.LoadInt32(&hitCount) != 2 {
		t.Fatalf("expected two requests due to retry, got %d", hitCount)
	}
}

func TestAnalyzeReturnsErrorWhenClientNotConfigured(t *testing.T) {
	client := NewClient(Config{})
	_, err := client.Analyze(context.Background(), domain.RiskDecisionEvent{})
	if err == nil {
		t.Fatalf("expected error when API key is missing")
	}
	if !strings.Contains(err.Error(), "not configured") {
		t.Fatalf("unexpected error: %v", err)
	}
}
