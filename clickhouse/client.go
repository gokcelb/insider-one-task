package clickhouse

import (
	"context"
	"fmt"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"

	"github.com/insider/event-ingestion/config"
)

type Client struct {
	conn driver.Conn
}

func NewClient(cfg config.ClickHouseConfig) (*Client, error) {
	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)},
		Auth: clickhouse.Auth{
			Database: cfg.Database,
			Username: cfg.Username,
			Password: cfg.Password,
		},
		Settings: clickhouse.Settings{
			"max_execution_time": 60,
		},
		Compression: &clickhouse.Compression{
			Method: clickhouse.CompressionLZ4,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to clickhouse: %w", err)
	}

	return &Client{conn: conn}, nil
}

func (c *Client) Conn() driver.Conn {
	return c.conn
}

func (c *Client) Ping(ctx context.Context) error {
	return c.conn.Ping(ctx)
}

func (c *Client) Close() error {
	return c.conn.Close()
}
