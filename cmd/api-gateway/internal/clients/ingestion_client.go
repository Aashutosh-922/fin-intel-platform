// package clients

// import (
// 	"bytes"
// 	"context"
// 	"encoding/json"
// 	"net/http"
// )

// type IngestionClient struct {
// 	baseURL string
// 	client  *http.Client
// }

// func NewIngestionClient() *IngestionClient {
// 	return &IngestionClient{
// 		baseURL: "http://ingestion:8080",
// 		client:  &http.Client{},
// 	}
// }

// type CreateTransactionRequest struct {
// 	UserID   string  `json:"user_id"`
// 	Amount   float64 `json:"amount"`
// 	Currency string  `json:"currency"`
// }

// func (c *IngestionClient) CreateTransaction(
// 	ctx context.Context,
// 	req CreateTransactionRequest,
// ) error {

// 	body, err := json.Marshal(req)
// 	if err != nil {
// 		return err
// 	}

// 	httpReq, err := http.NewRequestWithContext(
// 		ctx,
// 		http.MethodPost,
// 		c.baseURL+"/transactions",
// 		bytes.NewReader(body),
// 	)
// 	if err != nil {
// 		return err
// 	}

// 	httpReq.Header.Set("Content-Type", "application/json")

// 	resp, err := c.client.Do(httpReq)
// 	if err != nil {
// 		return err
// 	}
// 	defer resp.Body.Close()

// 	if resp.StatusCode >= 300 {
// 		return fmt.Errorf("ingestion returned %d", resp.StatusCode)
// 	}

// 	return nil
// }

package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"fmt"

	"github.com/Aashutosh-922/fin-intel-platform/cmd/api-gateway/internal/handlers"
)

type IngestionClient struct {
	baseURL string
	http    *http.Client
}

func NewIngestionClient(baseURL string) *IngestionClient {
	return &IngestionClient{
		baseURL: baseURL,
		http:    &http.Client{},
	}
}

func (c *IngestionClient) CreateTransaction(
	ctx context.Context,
	req handlers.CreateTransactionRequest,
) error {

	body, _ := json.Marshal(req)

	httpReq, _ := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.baseURL+"/transactions",
		bytes.NewBuffer(body),
	)

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("ingestion failed: %s", resp.Status)
	}

	return nil
}