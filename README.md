# Event Ingestion System

A high-throughput event ingestion backend using Go, ClickHouse, and Redpanda that handles ~2,000 events/second average (20,000 peak), with pre-aggregated metrics via materialized views.

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
  "campaign_id": "summer_sale_2024",
  "user_id": "user_12345",
  "timestamp": 1708732800000,
  "tags": ["electronics", "featured"],
  "metadata": {
    "product_id": "SKU-001",
    "page_url": "/products/laptop"
  }
}
```

**Response:**

- `202 Accepted`: `{"status": "accepted"}`
- `400 Bad Request`: `{"error": "validation error message"}`

**Validation Rules:**

- `event_name`: Required
- `user_id`: Required
- `timestamp`: Required, Unix milliseconds
- `channel`: Must be one of: web, mobile, api, email, push (or empty)

### GET /metrics

Query aggregated metrics.

**Query Parameters:**

- `event_name` (required): Event name to filter by
- `event_name` (required): Event name to filter by
- `from`: Unix timestamp in seconds (default: 24 hours ago)
- `to`: Unix timestamp in seconds (default: now)
- `group_by`: Supports "channel" or "hour"

**Example:**

```bash
curl "http://localhost:8080/metrics?event_name=product_view&from=1706745600&to=1706832000&group_by=channel"
```

**Response:**

```json
{
  "event_name": "product_view",
  "from": 1723400000,
  "to": 1723486400,
  "grouped_by": "channel",
  "total_events": 15420,
  "unique_users": 8905,
  "data": [
    {
      "group": "web",
      "total_events": 10200,
      "unique_users": 5100
    },
    {
      "group": "mobile_ios",
      "total_events": 3120,
      "unique_users": 2005
    },
    {
      "group": "mobile_android",
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

## Configuration

Environment variables:

| Variable               | Default         | Description                     |
| ---------------------- | --------------- | ------------------------------- |
| `SERVER_PORT`          | 8080            | HTTP server port                |
| `SERVER_READ_TIMEOUT`  | 5s              | HTTP read timeout               |
| `SERVER_WRITE_TIMEOUT` | 10s             | HTTP write timeout              |
| `KAFKA_BROKERS`        | localhost:19092 | Kafka/Redpanda broker addresses |
| `KAFKA_TOPIC`          | events          | Topic for events                |
| `CLICKHOUSE_HOST`      | localhost       | ClickHouse host                 |
| `CLICKHOUSE_PORT`      | 9000            | ClickHouse native port          |
| `CLICKHOUSE_DATABASE`  | events_db       | Database name                   |
| `CLICKHOUSE_USERNAME`  | default         | ClickHouse username             |
| `CLICKHOUSE_PASSWORD`  |                 | ClickHouse password             |

## Load Testing

The load test ramps up to 2,000 RPS sustained, then spikes to 20,000 RPS.

```bash
make load-test
```

Thresholds:

- p95 latency < 100ms
- p99 latency < 200ms
- Error rate < 1%

Results are saved to `loadtest/results.json`.

## Project Structure

```
├── cmd/
│   └── root.go                     # Entry point
├── clickhouse/
│   ├── migrations/                 # Database migrations
│   │   └── 001_initial_schema.up.sql
│   ├── repository/
│   │   └── metrics.go              # Metrics queries
│   ├── client.go                   # ClickHouse connection
│   └── migrate.go                  # Migration runner
├── events/
│   ├── handler.go                  # POST /events and /events/bulk handlers
│   ├── model.go                    # Event models
│   └── service.go                  # Event validation, Kafka publishing
├── metrics/
│   ├── handler.go                  # GET /metrics handler
│   ├── model.go                    # Metrics models
│   └── service.go                  # Metrics query logic
├── kafka/
│   └── producer.go                 # Redpanda producer (franz-go)
├── config/
│   └── config.go                   # Environment configuration
├── loadtest/
│   └── script.js                   # k6 load test script
├── docker-compose.yml              # Docker Compose configuration
├── Dockerfile
├── Makefile
└── README.md
```
