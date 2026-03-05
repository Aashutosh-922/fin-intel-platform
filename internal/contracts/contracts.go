package contracts

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

type orderEvent struct {
	EventID   string  `json:"event_id"`
	OrderID   string  `json:"order_id"`
	UserID    string  `json:"user_id"`
	Symbol    string  `json:"symbol"`
	Side      string  `json:"side"`
	Type      string  `json:"type"`
	Price     float64 `json:"price"`
	Quantity  int     `json:"quantity"`
	CreatedAt int64   `json:"created_at"`
}

type orderCancelEvent struct {
	OrderID string `json:"order_id"`
	Symbol  string `json:"symbol"`
}

type tradeEvent struct {
	TradeID     string  `json:"trade_id"`
	BuyOrderID  string  `json:"buy_order_id"`
	SellOrderID string  `json:"sell_order_id"`
	BuyUserID   string  `json:"buy_user_id"`
	SellUserID  string  `json:"sell_user_id"`
	Symbol      string  `json:"symbol"`
	Price       float64 `json:"price"`
	Quantity    int     `json:"quantity"`
	Timestamp   int64   `json:"timestamp"`
}

type bookEvent struct {
	Symbol    string      `json:"symbol"`
	Bids      []bookLevel `json:"bids"`
	Asks      []bookLevel `json:"asks"`
	Spread    float64     `json:"spread"`
	Timestamp int64       `json:"timestamp,omitempty"`
}

type marketTick struct {
	Symbol    string  `json:"symbol"`
	Price     float64 `json:"price"`
	Timestamp int64   `json:"timestamp"`
}

type marketAlert struct {
	Symbol     string  `json:"symbol"`
	Type       string  `json:"type"`
	ZScore     float64 `json:"z_score"`
	Volatility float64 `json:"volatility"`
	Timestamp  int64   `json:"timestamp"`
}

type riskDecision struct {
	EventID       string `json:"event_id"`
	TransactionID string `json:"transaction_id"`
	RiskScore     int    `json:"risk_score"`
	Decision      string `json:"decision"`
	Flagged       bool   `json:"flagged"`
	CreatedAt     int64  `json:"created_at"`
}

type bookLevel struct {
	Price    float64 `json:"price"`
	Quantity int     `json:"quantity"`
}

func ValidateTopicPayload(topic string, payload []byte) error {
	switch topic {
	case "orders":
		var e orderEvent
		if err := strictDecode(payload, &e); err != nil {
			return err
		}
		if e.OrderID == "" || e.UserID == "" || e.Symbol == "" || e.Side == "" || e.Type == "" || e.Quantity <= 0 {
			return errors.New("orders contract violation: required fields missing/invalid")
		}
	case "order-cancel":
		var e orderCancelEvent
		if err := strictDecode(payload, &e); err != nil {
			return err
		}
		if e.OrderID == "" || e.Symbol == "" {
			return errors.New("order-cancel contract violation: required fields missing")
		}
	case "trade-executed":
		var e tradeEvent
		if err := strictDecode(payload, &e); err != nil {
			return err
		}
		if e.Symbol == "" || e.Price <= 0 || e.Quantity <= 0 {
			return errors.New("trade-executed contract violation")
		}
	case "orderbook-snapshots", "orderbook-deltas":
		var e bookEvent
		if err := strictDecode(payload, &e); err != nil {
			return err
		}
		if e.Symbol == "" {
			return errors.New(topic + " contract violation: symbol required")
		}
	case "market-ticks":
		var e marketTick
		if err := strictDecode(payload, &e); err != nil {
			return err
		}
		if e.Symbol == "" || e.Price <= 0 {
			return errors.New("market-ticks contract violation")
		}
	case "market-alerts":
		var e marketAlert
		if err := strictDecode(payload, &e); err != nil {
			return err
		}
		if e.Symbol == "" || e.Type == "" {
			return errors.New("market-alerts contract violation")
		}
	case "risk-decisions":
		var e riskDecision
		if err := strictDecode(payload, &e); err != nil {
			return err
		}
		if e.TransactionID == "" || strings.TrimSpace(e.Decision) == "" {
			return errors.New("risk-decisions contract violation")
		}
	}
	return nil
}

func strictDecode(payload []byte, dst any) error {
	dec := json.NewDecoder(bytes.NewReader(payload))
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return fmt.Errorf("contract decode failed: %w", err)
	}
	return nil
}
