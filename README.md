# Event Ingestion System

A high-throughput event ingestion backend using Go, ClickHouse, and Redpanda that handles ~2,000 events/second average (20,000 peak), with queryable metrics data.

## Quick Start

```bash
# Start all services
make docker-up

# Check health
make health
make ready

# Send a test event
make test-event

# Query metrics
make test-metrics

# Run load tests
make load-test

# Stop services
make docker-down
```

## API Endpoints

### POST /events

Submit an event for ingestion.

**Request:**

```json
{
    "event_name": "product_view",
    "channel": "web",
    "campaign_id": "cmp_987",
    "user_id": "user_123",
    "timestamp": 1723475612,
    "tags": ["electronics", "homepage", "flash_sale"],
    "metadata": {
        "product_id": "prod-789",
        "price": 129.99,
        "currency": "TRY",
        "referrer": "google"
    }
}
```

**Response:**

- `202 Accepted`: `{"status": "accepted"}`
- `400 Bad Request`: `{"error": "validation error message"}`

**Validation Rules:**

- `event_name`: Required
- `user_id`: Required
- `timestamp`: Required, Unix seconds, must be in the past and positive
- `channel`: Required, must be one of: web, mobile, api, email, push

### GET /metrics

Query aggregated metrics.

**Query Parameters:**

- `event_name` (required): Event name to filter by
- `from`: Unix timestamp in seconds
- `to`: Unix timestamp in seconds
- `group_by`: Supports "channel", "hour", or "day"

**Example**

Only filtered by event name.

```bash
curl "http://localhost:8080/metrics?event_name=product_view"
```

**Response:**

```json
{
  "event_name": "product_view",
  "total_events": 10200,
  "unique_users": 5100
}
```

**Example:**

Filtered by event name and grouped by channel.

```bash
curl "http://localhost:8080/metrics?event_name=product_view&from=1772024670&to=1772024670&group_by=channel"
```

**Response:**

```json
{
  "event_name": "product_view",
  "from": 1772024670,
  "to": 1772024670,
  "grouped_by": "channel",
  "data": [
    {
      "group": "web",
      "total_events": 10200,
      "unique_users": 5100
    },
    {
      "group": "mobile",
      "total_events": 3120,
      "unique_users": 2005
    },
    {
      "group": "api",
      "total_events": 2100,
      "unique_users": 1800
    }
  ]
}
```

### GET /health

Basic health check.

**Response:** `{"status": "healthy"}`

### GET /ready

Readiness check (verifies ClickHouse and Kafka connectivity).

**Response:** `{"status": "ready"}` or `503` with error details.


## Load Testing

The load test ramps up to 2,000 RPS sustained, then spikes to 20,000 RPS.

```bash
make load-test
```

Thresholds:

- p95 latency < 60ms
- p99 latency < 150ms
- Error rate < 1%

Results are saved to `loadtest/results.json`.

Example result:

| Threshold | Result | Status |
| --- | --- | --- |
| p95 < 60ms | 48.5ms | Pass |
| p99 < 150ms | 72.5ms | Pass |
| Error rate < 1% | 0% | Pass |

## Technical Decision Making

### Tech Stack

**HTTP Framework:** Gin

Has built-in JSON binding and validation and has good performance.

**Database**: ClickHouse

ClickHouse is specifically built for analytical processing and thrives with large insertions of data at a time without any fine-tuning.

**Message Broker**: Redpanda

It has full on compatibility with Kafka, it's lightweight and easy to setup for local environment.

### Optimizations & Trade-offs

Below are the two important requirements that allow us to make certain optimizations.

> **R1** — *"Metrics endpoint does not need to be fully real-time."*

> **R2** — *"Ingestion should be near real-time and should not block under load."*

#### 1. Deduplication: eventual (database) vs. strict (application layer)

**Optimization:** I'm using `ReplacingMergeTree` for deduplication. ClickHouse merges duplicate rows in the background, which allows for faster inserts and faster `GET /metrics` queries.

**Trade-off:** Duplicate events may show up in metrics for a short time before the next background merge. This is fine because **R1** allows eventual consistency. Doing strict deduplication in the application layer (e.g. checking a Redis set or doing a `SELECT` before each `INSERT`) would avoid this but would add extra latency per event, which goes against **R2**.

#### 2. Kafka publish: synchronous vs. fire-and-forget

**Optimization:** I'm synchronously pushing events to the Kafka/Redpanda topic — the handler waits for the broker to acknowledge the write before returning `202 Accepted`. This way, accepted events are durable and won't be silently dropped if the ClickHouse consumer restarts.

**Trade-off:** Waiting for the broker acknowledgement is slower than fire-and-forget. But my load test results show it's still fast enough (average **~15.9 ms**, p95 **~40 ms**), so **R2** is satisfied in practice. The better choice here really depends on the durability requirements. If losing some events under extreme load is acceptable, fire-and-forget would be faster.

#### Ingestion flow: API → Kafka → ClickHouse (async batch)

1. **Validate & publish:** The API validates the incoming event and pushes it to the Kafka/Redpanda topic. Validation failures are rejected immediately with no broker write.
2. **Batch buffering:** Redpanda holds events until `kafka_max_block_size` (65,536 rows by default) or a flush interval is reached. Bulk inserts are much more efficient for ClickHouse than one-insert-per-event.
3. **Native Kafka Engine:** ClickHouse's built-in Kafka table engine consumes batches directly, eliminating the need for a custom consumer process.
4. **Query layer:** `GET /metrics` reads from the events table. Background merges handle deduplication over time, keeping query performance high.

## TODOs

Below are some TODOs which I would have implemented given more time, as well as some that are for production-grade apps.

- [ ] Structured logging
- [ ] Request logging middleware
- [ ] Unit and integration tests
- [ ] Authentication
- [ ] Better config management (e.g. Viper)
- [ ] Pagination on `GET /metrics`
- [ ] OpenAPI/Swagger documentation
- [ ] Rate limiting
- [ ] TLS
- [ ] Prometheus metrics
