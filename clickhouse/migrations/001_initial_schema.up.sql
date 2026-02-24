CREATE DATABASE IF NOT EXISTS events_db;

CREATE TABLE IF NOT EXISTS events_db.events (
    event_hash    UInt64,
    event_name    LowCardinality(String),
    channel       LowCardinality(String),
    campaign_id   String,
    user_id       String,
    timestamp     DateTime,
    tags          Array(String),
    metadata      String
)
ENGINE = ReplacingMergeTree()
PARTITION BY toYYYYMMDD(timestamp)
ORDER BY (event_name, timestamp, channel, event_hash);

CREATE TABLE IF NOT EXISTS events_db.events_kafka (
    event_hash    UInt64,
    event_name    String,
    channel       String,
    campaign_id   String,
    user_id       String,
    timestamp     UInt64,
    tags          Array(String),
    metadata      String
)
ENGINE = Kafka()
SETTINGS
    kafka_broker_list = 'redpanda:9092',
    kafka_topic_list = 'events',
    kafka_group_name = 'clickhouse_events_consumer',
    kafka_format = 'JSONEachRow',
    kafka_max_block_size = 65536;

CREATE MATERIALIZED VIEW IF NOT EXISTS events_db.events_kafka_mv
TO events_db.events AS
SELECT
    event_hash,
    event_name,
    channel,
    campaign_id,
    user_id,
    fromUnixTimestamp(timestamp) AS timestamp,
    tags,
    metadata
FROM events_db.events_kafka;
