# 🧠 Fin-Intel Platform

> **Event-Driven Fraud Detection & Financial Intelligence Platform**
> Built using Go, Kafka, TimescaleDB, Docker & CQRS architecture.

The Fin-Intel Platform is a fintech-grade distributed system designed to process real-time financial transactions, asynchronously evaluate fraud risk, and maintain an immutable, event-sourced timeline. The architecture is built for high throughput, fault tolerance, and time-series analytics.

## 🏗 High-Level Architecture

```text
                 ┌────────────────────┐
                 │    API Gateway     │
                 │  (Auth + Routing)  │
                 └─────────┬──────────┘
                           │
         ┌─────────────────┼─────────────────┐
         │                                   │
         ▼                                   ▼
  ┌──────────────┐                    ┌──────────────┐
  │ Ingestion    │                    │ Timeline API │
  │ Service      │                    │ (Read Model) │
  └──────┬───────┘                    └──────┬───────┘
         │                                   │
         ▼                                   ▼
     Kafka Topic                        TimescaleDB
   "transactions"                   (transaction_events)
         │                                   ▲
         ▼                                   │
  ┌──────────────┐                   ┌───────┴──────┐
  │ Risk Engine  │                   │ Timescale    │
  └──────┬───────┘                   │ Writer       │
         │                           └───────▲──────┘
         ▼                                   │
     Kafka Topic                             │
  "risk-decisions" ──────────────────────────┘
         │
         ▼ (Future Scope)
  ┌──────────────┐ 
  │  AI Service  │ 
  └──────┬───────┘
         │
         ▼
     Kafka Topic
    "ai-insights"

```

## ⚙️ Core Services (Current State)

1. **API Gateway**: Acts as the entry point. Enforces JWT authentication, role-based access control (USER, ANALYST, ADMIN), and routes traffic.
2. **Ingestion Service**: Stateless producer that accepts transaction requests and publishes immutable events to the `transactions` Kafka topic.
3. **Risk Engine**: Consumes transactions, applies rule-based fraud scoring (e.g., velocity checks, amount spikes, geo-anomalies), and publishes to `risk-decisions`.
4. **Timescale Writer**: The CQRS write model. Consumes risk decisions and persists immutable events into a TimescaleDB hypertable.
5. **Timeline Query (CQRS Read Model)**: Reads directly from TimescaleDB, returning chronological transaction histories and time-bucketed analytics.

## 🗄 Database Design (TimescaleDB)

Optimized for time-series data and CQRS event sourcing:

```sql
CREATE TABLE transaction_events (
    event_id TEXT NOT NULL,
    transaction_id TEXT NOT NULL,
    event_type TEXT NOT NULL,
    payload JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

SELECT create_hypertable('transaction_events','created_at');
CREATE INDEX idx_txn_time ON transaction_events (transaction_id, created_at DESC);

```

## ✅ Current Feature Status

| Feature | Status |
| --- | --- |
| Event-driven ingestion pipeline | ✅ Done |
| Kafka integration & async decoupling | ✅ Done |
| Deterministic Risk Evaluation Engine | ✅ Done |
| TimescaleDB event store (Hypertables) | ✅ Done |
| CQRS Read/Write separation | ✅ Done |
| JWT Auth + Role-Based Access Control | ✅ Done |
| Docker container orchestration | ✅ Done |

---

## 🚀 Future Scope Phase 1: AI Intelligence Layer

Moving from rule-based risk evaluation to probabilistic, explainable AI.

* **Dedicated AI Consumer**: A new Go microservice acting as a Kafka consumer listening to the `risk-decisions` topic.
* **Fraud Explainability**: When the risk engine flags a transaction, the AI service evaluates the timeline and payload to generate human-readable reasoning and confidence scores.
* **Event-Sourced Insights**: The AI service publishes to an `ai-insights` topic, which is then written to the TimescaleDB hypertable. The timeline becomes: `TRANSACTION_CREATED` → `RISK_EVALUATED` → `AI_ANALYSIS_GENERATED`.

## 💹 Future Scope Phase 2: Real-Time Stock Trading Pivot

Evolving the platform from a generic transaction processor into a real-time stock trading backend.

* **Domain Shift**: Refactoring the API to ingest `StockOrder` models (Symbol, Order Type, Limit Price, Quantity) instead of generic transactions.
* **Margin & Exposure Engine**: Upgrading the Risk Engine to evaluate user margin limits and exposure limits before routing to the market.
* **Market Data Hypertable**: Introducing a `market_ticks` TimescaleDB hypertable to store incoming price changes and volume ticks.
* **Order Matching Engine**: A new service to consume validated buy/sell orders and match them based on price-time priority, emitting `trade-executed` events.
* **AI Market Analytics**: Utilizing the AI service to consume Timescale time buckets and trade velocity to predict short-term volatility and detect abnormal order clusters.

## 🧪 Local Development

Clone the repository and spin up the distributed environment using Docker:

```bash
git clone https://github.com/Aashutosh-922/fin-intel-platform.git
cd fin-intel-platform

# Build and start all microservices, Kafka, Zookeeper, and TimescaleDB
docker compose up --build

```

**Test the flow:**

1. `POST /transactions`
2. `GET /transactions/{id}/timeline`
