package config

import (
	"fmt"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Server     ServerConfig
	Kafka      KafkaConfig
	ClickHouse ClickHouseConfig
}

type ServerConfig struct {
	Port         int           `envconfig:"SERVER_PORT" default:"8080"`
	ReadTimeout  time.Duration `envconfig:"SERVER_READ_TIMEOUT" default:"5s"`
	WriteTimeout time.Duration `envconfig:"SERVER_WRITE_TIMEOUT" default:"10s"`
}

type KafkaConfig struct {
	Brokers []string `envconfig:"KAFKA_BROKERS" default:"localhost:19092"`
	Topic   string   `envconfig:"KAFKA_TOPIC" default:"events"`
}

type ClickHouseConfig struct {
	Host     string `envconfig:"CLICKHOUSE_HOST" default:"localhost"`
	Port     int    `envconfig:"CLICKHOUSE_PORT" default:"9000"`
	Database string `envconfig:"CLICKHOUSE_DATABASE" default:"events_db"`
	Username string `envconfig:"CLICKHOUSE_USERNAME" default:"default"`
	Password string `envconfig:"CLICKHOUSE_PASSWORD" default:""`
}

func Load() (*Config, error) {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, fmt.Errorf("failed to process env config: %w", err)
	}
	return &cfg, nil
}
