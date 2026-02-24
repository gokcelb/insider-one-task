# TECHNICAL DECISION MAKING

## Tech Stack

**Programming language:** Golang

It is the language used in your company and the dominant language for high-throughput backend systems because of its compiled speed and lightweight Goroutines.

**HTTP Framework:** Gin

Has built-in JSON binding and validation and has good performance.

**Database**: ClickHouse

ClickHouse is specifically built for analytical processing and thrives with large insertions of data at a time without any fine-tuning.

**Message Broker**: Redpanda

I chose Redpanda because although it has full on compatibility with Kafka, it's lightweight and easy to setup for local environment.

## Optimizations

Below are the two important API specifications that allow us to make certain optimizations.

### 1. Metrics endpoint does not need to be fully real-time

This allows us to optimize API response times by allowing ClickHouse's background deduplication and speeding up our metrics queries.

### 2. Ingestion should be near real-time and should not block under load

In order not to block the server under load and have low latency, we need to separate the API layer from the database layer. To do this, I introduced a message broker (Redpanda) to queue the events and process them asynchronously.

With this, here is how the end-to-end flow looks like:

1. The API receives the request and the data received is validated. If valid, the event is pushed to the message broker asynchronously and a success response is returned. If invalid, the event is not pushed to the queue, and an error response is returned.
2. The message broker holds onto the event data until a certain batch size is reached, this allows us to process them in the background in batches. Batch processing is good for multiple reasons:
   a. Prevents congestion in the API layer and allows quick response times
   b. Allows us to correctly use ClickHouse's insertion mechanism as Clickhouse can get overwhelmed with high single insert requests, while it thrives with bulk inserts.
3. There is a ClickHouse Kafka Engine that we can utilize. It has built-in `kafka_max_block_size` configuration for the "big batch insertions" synergy with ClickHouse. As we don't need to do any further processing on the event data before inserting to the database, ClickHouse's native Kafka engine saves us from writing a custom consumer.
4. `GET /metrics` endpoint queries events table that has eventual deduplication to serve event metrics quickly.

## Out of Scope / Future Work

- [ ] Structured logging
- [ ] Request logging middleware
- [ ] Unit and integration tests
- [ ] Rate limiting
- [ ] TLS
- [ ] Authentication
- [ ] Prometheus metrics
- [ ] Pagination on `GET /metrics`
- [ ] OpenAPI/Swagger documentation
