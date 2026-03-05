# Operations Runbook

This runbook contains practical incident response playbooks for the Fin Intel Platform.

## 1) Quick Health Checklist

Run from `fin-intel-platform/docker`.

```bash
docker compose ps -a
curl -sS http://localhost:8080/health
curl -sS http://localhost:8081/health
curl -sS http://localhost:8082/health
curl -sS http://localhost:8070/health
curl -sS http://localhost:8071/health
```

If any critical service is down (`kafka`, `ingestion`, `risk-engine`, `timescale-writer`, `api-gateway`), jump to the relevant playbook below.

---

## 2) Incident: Kafka Down / Crashing

### Symptoms
- Gateway returns `502 ingestion failed`.
- Ingestion logs show:
  - `lookup kafka ... no such host`
  - `Unknown Topic Or Partition`
  - connection refused to `kafka:29092`
- `docker-kafka-1` is `Exited` in `docker compose ps -a`.

### Diagnose
```bash
docker compose ps -a | rg kafka
docker logs --tail 200 docker-kafka-1
df -h
```

### Recovery
1. Free disk if root is full:
```bash
docker builder prune -f
docker image prune -f
```
2. Restart Kafka:
```bash
docker start docker-kafka-1
```
3. Recreate/check topics:
```bash
docker compose up -d --no-deps kafka-init
docker logs --tail 120 docker-kafka-init-1
```
4. Verify broker + topics:
```bash
docker ps --filter name=docker-kafka-1 --format '{{.Names}} {{.Status}}'
docker exec docker-kafka-1 kafka-topics --bootstrap-server kafka:29092 --list | sort
```

### Post-Recovery Verification
```bash
curl -i -X POST http://localhost:8080/transactions \
  -H 'Authorization: Bearer test' \
  -H 'X-Role: USER' \
  -H 'Content-Type: application/json' \
  -d '{"user_id":"runbook-user","amount":100.5,"currency":"USD"}'
```
Expect `202 Accepted`.

---

## 3) Incident: Topics Missing After Kafka Restart

### Symptoms
- Producers fail with `Unknown Topic Or Partition`.

### Recovery
```bash
docker compose up -d --no-deps kafka-init
docker logs --tail 200 docker-kafka-init-1
```

### Verify Required Topics
```bash
docker exec docker-kafka-1 kafka-topics --bootstrap-server kafka:29092 --list | sort
```
Expected set:
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

---

## 4) Incident: Disk Full (`/` at 100%)

### Symptoms
- Kafka exits with `No space left on device`.
- Build/deploy instability.
- Containers restart frequently.

### Diagnose
```bash
df -h
docker system df
```

### Fast Cleanup
```bash
docker builder prune -f
docker image prune -f
```

### Optional Deeper Cleanup (careful)
```bash
docker container prune -f
docker volume prune -f
```
Only run if you accept deleting stopped containers/unused volumes.

### Verify
```bash
df -h
```
Target: maintain safe free space buffer before load tests.

---

## 5) Incident: Ingestion Returns 500

### Symptoms
- `POST /transactions` returns 500 or gateway returns 502.

### Diagnose
```bash
docker logs --tail 200 docker-ingestion-1
```
Check for:
- DB connection failures
- Kafka publish failures
- DNS or topic errors

### Recovery
- If Kafka issue: use Kafka playbook.
- If Postgres issue: use Postgres playbook.
- Restart ingestion after dependency recovery:
```bash
docker compose up -d --no-deps ingestion
```

---

## 6) Incident: Postgres Unavailable / Migration Failures

### Symptoms
- Ingestion or portfolio service cannot start/connect.
- Errors around `failed to connect to postgres`.

### Diagnose
```bash
docker compose ps -a | rg postgres
docker logs --tail 200 docker-postgres-1
docker logs --tail 200 docker-postgres-init-1
```

### Recovery
```bash
docker compose up -d postgres
sleep 3
docker compose up -d --no-deps postgres-init
```

### Verify
```bash
docker exec -e PGPASSWORD=fintech docker-postgres-1 \
  psql -U fintech -d fintech -c '\dt'
```
Expect at least: `transactions`, `positions`, `market_prices`.

---

## 7) Incident: Timescale Writer / Timeline Missing

### Symptoms
- `/transactions/{id}/timeline` returns empty or not found.
- `/ai/query` returns `ai insight not found`.

### Diagnose
```bash
docker logs --tail 200 docker-timescale-writer-1
```

Check recent persisted events:
```bash
docker exec -e PGPASSWORD=timescale docker-timescaledb-1 \
  psql -U timescale -d events \
  -c "select transaction_id,event_type,created_at from transaction_events order by created_at desc limit 20;"
```

### Recovery
```bash
docker compose up -d --no-deps timescale-writer
```

### Verify AI persistence path
- Confirm both `APPROVED` and `AI_ANALYSIS` rows exist for a transaction.

---

## 8) Incident: API Gateway Read Path Wrong/Empty

### Symptoms
- `/transactions/{id}` missing risk/explanation context.
- `/ai/query` gives stale or missing data.

### Diagnose
```bash
docker logs --tail 200 docker-api-gateway-1
```

Validate underlying DB records:
```bash
# Postgres transaction row
docker exec -e PGPASSWORD=fintech docker-postgres-1 \
  psql -U fintech -d fintech \
  -c "select id,user_id,amount,status,created_at from transactions order by created_at desc limit 5;"

# Timescale timeline rows
docker exec -e PGPASSWORD=timescale docker-timescaledb-1 \
  psql -U timescale -d events \
  -c "select transaction_id,event_type,created_at from transaction_events order by created_at desc limit 10;"
```

### Recovery
```bash
docker compose up -d --no-deps api-gateway
```

---

## 9) Incident: Order Placement Rejected Unexpectedly

### Symptoms
- `400 duplicate order`
- `400 missing required order fields`

### Notes
- `order_id` is idempotent key; must be unique per new order.
- Payload must be snake_case fields.

Valid payload example:
```json
{
  "order_id": "ord-rb-1",
  "user_id": "user-rb-1",
  "symbol": "INFY",
  "side": "BUY",
  "type": "LIMIT",
  "price": 1500,
  "quantity": 5
}
```

### Recovery
- Retry with a new `order_id` and valid fields.
- If persistent, check:
```bash
docker logs --tail 200 docker-order-service-1
```

---

## 10) Incident: Portfolio Not Updating

### Symptoms
- Trades appear in `trade-executed` but positions don’t change.

### Diagnose
```bash
docker logs --tail 200 docker-portfolio-service-1

docker exec -e PGPASSWORD=fintech docker-postgres-1 \
  psql -U fintech -d fintech \
  -c "select user_id,symbol,quantity,avg_price,realized_pnl,unrealized_pnl from positions order by symbol,user_id;"
```

### Verify topic side
```bash
docker exec docker-kafka-1 kafka-console-consumer \
  --bootstrap-server kafka:29092 \
  --topic trade-executed --from-beginning --max-messages 5 --timeout-ms 4000
```

---

## 11) Incident: Volatility Alerts Missing

### Symptoms
- `/alerts` remains empty when expected spikes should be detected.

### Important detail
- Detector needs enough history (`window` size) before generating spikes.

### Diagnose
```bash
docker logs --tail 200 docker-volatility-ai-1
curl -sS http://localhost:8071/alerts
```

### Spike simulation (synthetic)
```bash
docker exec docker-kafka-1 bash -lc '\
(for i in $(seq 1 29); do \
  echo "{\"symbol\":\"VOLTEST\",\"price\":100,\"quantity\":1,\"timestamp\":1772185000}"; \
done; \
echo "{\"symbol\":\"VOLTEST\",\"price\":1000,\"quantity\":1,\"timestamp\":1772185001}") \
| kafka-console-producer --bootstrap-server kafka:29092 --topic trade-executed'
```

Then:
```bash
curl -sS http://localhost:8071/alerts
```

---

## 12) Incident: AI Service Not Producing Gemini Insights

### Symptoms
- `ai-insights` topic stays empty while `risk-decisions` has traffic.
- `docker-ai-service-1` logs contain:
  - `gemini client not configured`
  - `gemini status=404`
  - repeated `analysis error`

### Diagnose
```bash
docker logs --tail 200 docker-ai-service-1
docker exec docker-kafka-1 kafka-console-consumer \
  --bootstrap-server kafka:29092 \
  --topic risk-decisions --from-beginning --max-messages 5 --timeout-ms 4000
```

Check runtime env in compose:
```bash
docker compose exec ai-service env | rg 'AI_MODE|AI_PROVIDER|GEMINI_API_KEY|GEMINI_MODEL|GEMINI_BASE_URL'
```

### Recovery
1. Set envs in `docker/.env`:
```env
AI_MODE=hybrid
AI_PROVIDER=gemini
GEMINI_API_KEY=<your-key>
GEMINI_BASE_URL=https://generativelanguage.googleapis.com/v1beta
```
2. Restart only AI service:
```bash
docker compose up -d --no-deps ai-service
```
3. Validate service package health from repo root:
```bash
GOCACHE=/tmp/go-build go test ./cmd/ai-service/...
```

### Post-Recovery Verification
- `docker logs --tail 200 docker-ai-service-1` shows normal consume/publish loop.
- `transaction_events` includes `AI_ANALYSIS` rows for newly processed transactions.

---

## 13) Controlled Restart Order

When the environment is unstable, restart in dependency order:

1. `zookeeper`
2. `kafka`
3. `kafka-init`
4. `postgres`, `timescaledb`
5. `postgres-init`, `timescaledb-init`
6. core consumers/producers (`ingestion`, `risk-engine`, `ai-service`, `timescale-writer`)
7. trading path (`order-service`, `matching-engine`, `portfolio-service`, `volatility-ai`, `market-data-service`)
8. `api-gateway`

Example:
```bash
docker compose up -d zookeeper kafka
sleep 3
docker compose up -d kafka-init postgres timescaledb
sleep 3
docker compose up -d postgres-init timescaledb-init
sleep 3
docker compose up -d ingestion risk-engine ai-service timescale-writer
sleep 3
docker compose up -d order-service matching-engine portfolio-service volatility-ai market-data-service
sleep 3
docker compose up -d api-gateway
```

---

## 14) Incident: WebSocket Market Stream Missing Data

### Symptoms
- `ws://localhost:8070/ws/market` connects but no messages.
- Messages arrive for wrong symbol/topic.

### Diagnose
```bash
docker logs --tail 200 docker-market-data-service-1
curl -sS http://localhost:8070/health
```

Verify key upstream topics have records:
```bash
docker exec docker-kafka-1 kafka-console-consumer \
  --bootstrap-server kafka:29092 --topic orderbook-deltas \
  --max-messages 3 --timeout-ms 5000

docker exec docker-kafka-1 kafka-console-consumer \
  --bootstrap-server kafka:29092 --topic market-alerts \
  --max-messages 3 --timeout-ms 5000

docker exec docker-kafka-1 kafka-console-consumer \
  --bootstrap-server kafka:29092 --topic risk-decisions \
  --max-messages 3 --timeout-ms 5000
```

### Correct subscription examples
- all streams: `ws://localhost:8070/ws/market`
- INFY ticks only: `ws://localhost:8070/ws/market?symbols=INFY&topics=market-ticks`
- orderbook deltas only: `ws://localhost:8070/ws/market?topics=orderbook-deltas`
- risk rejections only: `ws://localhost:8070/ws/market?topics=risk-rejections`

### Recovery
```bash
docker compose up -d --no-deps market-data-service
docker logs --tail 120 docker-market-data-service-1
```

---

## 15) Verification Commands (Golden Path)

1) Transaction create:
```bash
curl -i -X POST http://localhost:8080/transactions \
  -H 'Authorization: Bearer test' \
  -H 'X-Role: USER' \
  -H 'Content-Type: application/json' \
  -d '{"user_id":"golden-user","amount":1234.56,"currency":"USD"}'
```

2) Get latest transaction id:
```bash
docker exec -e PGPASSWORD=fintech docker-postgres-1 \
  psql -U fintech -d fintech -t -A \
  -c "select id from transactions order by created_at desc limit 1;"
```

3) Timeline row check:
```bash
docker exec -e PGPASSWORD=timescale docker-timescaledb-1 \
  psql -U timescale -d events \
  -c "select transaction_id,event_type,created_at from transaction_events order by created_at desc limit 10;"
```

4) AI query:
```bash
curl -i -X POST http://localhost:8080/ai/query \
  -H 'Authorization: Bearer test' \
  -H 'X-Role: ANALYST' \
  -H 'Content-Type: application/json' \
  -d '{"transaction_id":"<TX_ID>","question":"why approved?"}'
```

---

## 16) Escalation Guidance

Escalate to code-level investigation when:
- all infra is healthy,
- topics exist,
- DB connectivity is stable,
- but flow still fails for valid payloads.

Minimum artifacts to capture before escalation:
- `docker compose ps -a`
- tail logs of failing services
- last successful and failed request payloads
- `df -h` and `docker system df`

---

## 17) Preventive Disk Guard (Recommended)

To reduce repeat Kafka crashes from low root disk:

1) Keep aggressive Kafka retention active (already configured in `docker-compose.yml`):
- broker defaults (`KAFKA_LOG_RETENTION_HOURS`, `KAFKA_LOG_RETENTION_BYTES`)
- topic-specific retention applied by `kafka-init`

2) Run periodic Docker cleanup with:
```bash
bash docker/ops/disk-guard.sh
```

Optional cron (every 15 minutes):
```bash
*/15 * * * * cd /path/to/fin-intel-platform && bash docker/ops/disk-guard.sh >> /tmp/fin-intel-disk-guard.log 2>&1
```

Tune threshold:
```bash
THRESHOLD=80 bash docker/ops/disk-guard.sh
```
