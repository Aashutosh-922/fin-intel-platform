# Fin Intel Platform

Real-time, event-driven fintech platform for transaction ingestion, risk scoring, AI analysis, order matching, portfolio updates, and market volatility alerting.

This repository runs as a multi-service system on Docker Compose using Kafka for async communication, Postgres for operational state, and TimescaleDB for event timeline/history.

## What The System Is Capable Of

- Accepts transactions over HTTP and publishes canonical transaction events.
- Performs real-time risk evaluation with retry + DLQ handling.
- Generates AI risk insights from risk decisions.
- Stores decision and AI timeline events in TimescaleDB (hypertable).
- Accepts orders, emits order events, and performs matching in-memory.
- Emits trade executions and updates user portfolio positions.
- Consumes market ticks and updates unrealized PnL.
- Detects volatility spikes from trade streams and exposes alerts.
- Exposes a role-protected API gateway for primary read/write endpoints.
- Optional monitoring stack (Prometheus + Grafana + JMX exporter) via profile.

---

## High-Level Architecture

```text
Clients / Postman
       |
       v
+----------------+
|   API Gateway  |  :8080
+----------------+
   | create tx                        | ai query / reads
   v                                  v
+----------------+              +---------------------+
| Ingestion Svc  | :8081        | Postgres + Timescale|
+----------------+              +---------------------+
   | publish transactions                  ^
   v                                       |
+----------------+                         |
|     Kafka      | <-----------------------+
+----------------+
   |  risk-decisions     | ai-insights
   v                     v
+----------------+   +----------------+
|  Risk Engine   |   |   AI Service   |
+----------------+   +----------------+
          \             /
           \           /
            v         v
          +------------------+
          | Timescale Writer |
          +------------------+
                 |
                 v
          TimescaleDB timeline

Order Flow:
Order Service (:8082) -> Kafka(orders) -> Matching Engine (:8090)
-> Kafka(trade-executed) -> Portfolio Service + Volatility AI (:8071)

Market Flow:
Market Data Service -> Kafka(market-ticks) -> Portfolio Service
```

---

## Core Services

### 1) `api-gateway` (`cmd/api-gateway`)

Responsibilities:
- Public entrypoint for API calls.
- Auth + RBAC middleware.
- Routes transaction creation to ingestion service.
- Serves transaction reads, timeline reads, and AI query responses.

Important behavior:
- Requires `Authorization` header for protected routes.
- Uses `X-Role` for role checks (`USER`, `ANALYST`, `ADMIN`).
- Reads transaction state from Postgres and timeline/AI data from TimescaleDB.

### 2) `ingestion-service` (`cmd/ingestion-service`)

Responsibilities:
- Accepts `/transactions`.
- Validates and stores transaction in Postgres (`transactions` table).
- Publishes event to Kafka topic `transactions` (with retry + DLQ logic).

### 3) `risk-engine` (`cmd/risk-engine`)

Responsibilities:
- Consumes `transactions` and `transactions-retry`.
- Applies rules/scoring and outputs risk decision.
- Publishes to Kafka topic `risk-decisions`.
- Retries transient failures; sends poison messages to `transactions-dlq`.

### 4) `ai-service` (`cmd/ai-service`)

Responsibilities:
- Consumes `risk-decisions`.
- Produces AI analysis to topic `ai-insights`.
- Uses deterministic feature scoring (behavioral, velocity, geo) as baseline.
- Supports runtime modes:
  - `AI_MODE=deterministic` (binary default, no external dependency)
  - `AI_MODE=hybrid` (try external LLM, fallback to deterministic)
  - `AI_MODE=llm_only` (external LLM required)
- Docker Compose default is `AI_MODE=hybrid`.

Provider selection:
- `AI_PROVIDER=gemini` (default) or `AI_PROVIDER=openai`

Gemini envs (when using `hybrid` / `llm_only`):
- `GEMINI_API_KEY`
- `GEMINI_MODEL` (default `gemini-1.5-flash`)
- `GEMINI_BASE_URL` (default `https://generativelanguage.googleapis.com/v1beta`; root URL is auto-normalized to `/v1beta`)
- `GEMINI_TIMEOUT_MS`
- `GEMINI_MAX_RETRIES`

OpenAI envs (optional if `AI_PROVIDER=openai`):
- `OPENAI_API_KEY`
- `OPENAI_MODEL`
- `OPENAI_BASE_URL`
- `OPENAI_TIMEOUT_MS`
- `OPENAI_MAX_RETRIES`

### 5) `timescale-writer` (`cmd/timescale-writer`)

Responsibilities:
- Consumes `risk-decisions` and `ai-insights`.
- Persists normalized timeline records in `transaction_events` hypertable.
- Stores:
  - risk events as `event_type = APPROVED|REJECTED|...`
  - AI events as `event_type = AI_ANALYSIS`

### 6) `order-service` (`cmd/order-service`)

Responsibilities:
- Accepts order creation and order cancel requests.
- Validates payload (snake_case fields) + idempotency.
- Publishes to:
  - `orders` (create)
  - `order-cancel` (cancel)

### 7) `matching-engine` (`cmd/matching-engine`)

Responsibilities:
- Consumes `orders` and `order-cancel`.
- Maintains in-memory order books.
- Produces trade executions to `trade-executed`.
- Exposes REST snapshot + WebSocket streams.

### 8) `portfolio-service` (`cmd/portfolio-service`)

Responsibilities:
- Consumes `trade-executed` to update positions.
- Consumes `market-ticks` to compute unrealized PnL.
- Stores results in `positions` table.

### 9) `market-data-service` (`cmd/market-data-service`)

Responsibilities:
- Generates synthetic ticks for configured symbols.
- Publishes to `market-ticks` every second.
- Consumes `trade-executed`, `orderbook-snapshots`, `orderbook-deltas`, `market-alerts`, `risk-decisions`.
- Broadcasts market stream over WebSocket:
  - `GET /health` (HTTP, port `8070`)
  - `GET /ws/market` (WebSocket, port `8070`)
    - optional query params:
      - `symbols=INFY,TCS`
      - `topics=market-ticks,trade-executed,orderbook-snapshots,orderbook-deltas,market-alerts,risk-rejections`

### 10) `volatility-ai` (`cmd/volatility-ai`)

Responsibilities:
- Consumes `trade-executed`.
- Detects volatility spikes (z-score based window detector).
- Publishes alerts to `market-alerts`.
- Exposes:
  - `/health`
  - `/alerts`

---

## Kafka Topic Topology

Created by `kafka-init`:

- `transactions`
- `transactions-retry`
- `transactions-dlq`
- `risk-decisions`
- `ai-insights`
- `orders`
- `order-cancel`
- `trade-executed`
- `orderbook-snapshots`
- `orderbook-deltas`
- `market-alerts`
- `market-ticks`

Typical producers/consumers:

- `ingestion` -> `transactions`
- `risk-engine` consumes `transactions`, `transactions-retry`; produces `risk-decisions`
- `ai-service` consumes `risk-decisions`; produces `ai-insights`
- `timescale-writer` consumes `risk-decisions`, `ai-insights`
- `order-service` -> `orders`, `order-cancel`
- `matching-engine` consumes `orders`, `order-cancel`; produces `trade-executed`
- `matching-engine` also produces `orderbook-snapshots`, `orderbook-deltas`
- `portfolio-service` consumes `trade-executed`, `market-ticks`
- `volatility-ai` consumes `trade-executed`; produces `market-alerts`
- `market-data-service` consumes `trade-executed`, `orderbook-snapshots`, `orderbook-deltas`, `market-alerts`, `risk-decisions`; produces `market-ticks`; broadcasts filtered streams on `/ws/market`

---

## Data Schemas

### Postgres (`fintech` DB)

From `docker/db-init/postgres-init.sql`:

#### `transactions`
- `id UUID PRIMARY KEY`
- `user_id TEXT`
- `amount NUMERIC`
- `currency TEXT`
- `status TEXT` (default `RECEIVED`)
- `created_at TIMESTAMPTZ`

#### `positions`
- `user_id TEXT`
- `symbol TEXT`
- `quantity INT`
- `avg_price DOUBLE PRECISION`
- `realized_pnl DOUBLE PRECISION`
- `unrealized_pnl DOUBLE PRECISION`
- PK: `(user_id, symbol)`

#### `market_prices`
- `symbol TEXT PRIMARY KEY`
- `price DOUBLE PRECISION`
- `updated_at BIGINT`

### TimescaleDB (`events` DB)

From `docker/db-init/timescaledb-init.sql`:

#### `transaction_events` (hypertable on `created_at`)
- `event_id TEXT`
- `transaction_id TEXT`
- `event_type TEXT`
- `payload JSONB`
- `created_at TIMESTAMPTZ`

Index:
- `(transaction_id, created_at DESC)`

---

## API Endpoints

### API Gateway (`:8080`)

Public:
- `GET /health`

Protected (`Authorization` required):
- `POST /transactions` (`USER`, `ADMIN`)
- `GET /transactions/{id}` (`ANALYST`, `ADMIN`)
- `GET /transactions/{id}/timeline` (`ANALYST`, `ADMIN`)
- `POST /ai/query` (`ANALYST`, `ADMIN`)

### Ingestion (`:8081`)
- `GET /health`
- `POST /transactions`

### Order Service (`:8082`)
- `GET /health`
- `POST /orders`
- `POST /orders/{order_id}/cancel?symbol=...`

### Matching Engine (`:8090`)
- `GET /orderbook/{symbol}`
- `GET /ws/orderbook/{symbol}`
- `GET /ws/trades/{symbol}`

### Volatility AI (`:8071`)
- `GET /health`
- `GET /alerts`

---

## Canonical Flow Walkthroughs

### A) Transaction -> Risk -> AI -> Timeline

1. Client posts transaction to gateway.
2. Gateway forwards to ingestion.
3. Ingestion inserts into `transactions` and publishes `transactions` topic.
4. Risk engine consumes, scores, and publishes `risk-decisions`.
5. AI service consumes `risk-decisions` and publishes `ai-insights`.
6. Timescale writer consumes both topics and writes timeline rows.
7. Gateway reads from Postgres + Timescale for `/transactions/{id}` and `/timeline`.
8. `/ai/query` reads latest `AI_ANALYSIS` payload for `transaction_id`.

### B) Order -> Match -> Trade -> Portfolio

1. Client posts BUY/SELL orders to order service.
2. Order service publishes `orders`.
3. Matching engine matches and emits `trade-executed`.
4. Portfolio service updates `positions` for buyer/seller.

### C) Market Tick -> Unrealized PnL

1. Market data service emits ticks to `market-ticks`.
2. Portfolio service consumes and recalculates `unrealized_pnl`.
3. Updated values persist in `positions`.

### D) Trade -> Volatility Alerts

1. Volatility AI consumes `trade-executed`.
2. Detector computes z-score/volatility over rolling window.
3. High spike emits `VOLATILITY_SPIKE` alert and stores in-memory list.
4. Alert visible at `/alerts`.

---

## Running The System

From `docker/`:

```bash
cp .env.example .env
```

Set `GEMINI_API_KEY` in `.env` before starting if you want live Gemini calls.

Then run:

```bash
docker compose up -d --build
```

Core endpoints:
- Gateway: `http://localhost:8080`
- Ingestion: `http://localhost:8081`
- Order service: `http://localhost:8082`
- Matching engine: `http://localhost:8090`
- Volatility AI: `http://localhost:8071`

Optional monitoring profile:

```bash
docker compose --profile monitoring up -d
```

- Prometheus: `http://localhost:9090`
- Grafana: `http://localhost:3000`

---

## AI Service Verification

Compile + test AI service package:

```bash
GOCACHE=/tmp/go-build go test ./cmd/ai-service/...
```

Quick runtime smoke check (Gemini provider):

```bash
docker compose exec -e AI_MODE=hybrid -e AI_PROVIDER=gemini ai-service env | rg 'AI_MODE|AI_PROVIDER|GEMINI_MODEL|GEMINI_BASE_URL'
docker compose logs --tail 120 ai-service
```

Expected:
- startup log includes `ai-service consumer started`
- no recurring `gemini status=...` or `gemini client not configured` errors when `GEMINI_API_KEY` is set

---

## Postman

Import:
- `postman/fin-intel-platform-e2e.postman_collection.json`

The collection includes ordered requests for:
- health checks
- transaction + read/timeline
- ai query
- order placement/cancel
- orderbook
- volatility endpoints

Run with Newman:

```bash
newman run postman/fin-intel-platform-e2e.postman_collection.json
```

Note:
- WebSocket steps are Newman-compatible checks (URL/filter validation + market-data health), since Newman does not execute live WS subscriptions.

---

## Operational Notes / Troubleshooting

### 1) Kafka crash due to low disk space

Observed behavior:
- Kafka exits with `No space left on device`.
- Upstream services fail with DNS/connection/topic errors.

Symptoms:
- ingestion logs: `lookup kafka... no such host`, `Unknown Topic Or Partition`
- gateway returns `502 ingestion failed`

Recovery runbook:
1. Free disk (`docker builder prune -f`, remove unused images/containers).
2. Restart Kafka container.
3. Re-run `kafka-init` to recreate topics if needed.
4. Retry ingestion and confirm events flow.

### 2) Topic missing after Kafka restart

If producers return `Unknown Topic Or Partition`, re-run:

```bash
docker compose up -d --no-deps kafka-init
```

### 3) AI query not found

`POST /ai/query` returns `404 ai insight not found` when `AI_ANALYSIS` has not yet been persisted for that `transaction_id`.

### 4) Gemini provider misconfigured

Typical symptoms:
- ai-service logs show `gemini client not configured` (missing key)
- ai-service logs show `gemini status=404` (wrong base URL/version path)

Quick checks:

```bash
docker compose exec ai-service env | rg 'AI_MODE|AI_PROVIDER|GEMINI_API_KEY|GEMINI_BASE_URL|GEMINI_MODEL'
docker compose logs --tail 200 ai-service
```

Fix:
- Set `GEMINI_API_KEY` in `docker/.env`
- Prefer `GEMINI_BASE_URL=https://generativelanguage.googleapis.com/v1beta`

---

## Security/Design Notes

- Auth is header-based middleware (`Authorization` + `X-Role`) for local/dev simplicity.
- No TLS/JWT issuance in this repo yet.
- Matching engine state is in-memory (not durable).
- Some services are event-only and intentionally have no REST APIs.

---

## Suggested Next Improvements

- Introduce persistent order book and replay-safe recovery.
- Add OpenTelemetry traces across HTTP + Kafka hops.
- Add contract tests for topic payload schemas.
- Add retention/archival policy for `transaction_events`.
- Harden auth to JWT verification with signed claims.
