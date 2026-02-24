package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

type MetricsRepository struct {
	conn driver.Conn
}

type MetricRow struct {
	EventName   string `json:"event_name"`
	GroupKey    string `json:"group_key"`
	TotalCount  uint64 `json:"total_count"`
	UniqueUsers uint64 `json:"unique_users"`
}

func NewMetricsRepository(conn driver.Conn) *MetricsRepository {
	return &MetricsRepository{conn: conn}
}

func (r *MetricsRepository) GetMetrics(ctx context.Context, eventName string, startTime, endTime *time.Time, groupBy string) ([]MetricRow, error) {
	groupCol := "channel"
	if groupBy == "hour" {
		groupCol = "toString(toStartOfHour(timestamp))"
	}

	query := fmt.Sprintf(`
		SELECT
			event_name,
			%s AS group_key,
			count() AS total_count,
			uniq(user_id) AS unique_users
		FROM events_db.events
		WHERE event_name = @eventName
	`, groupCol)

	args := []any{
		driver.NamedValue{Name: "eventName", Value: eventName},
	}

	if startTime != nil {
		query += " AND timestamp >= @startTime"
		args = append(args, driver.NamedValue{Name: "startTime", Value: *startTime})
	}

	if endTime != nil {
		query += " AND timestamp <= @endTime"
		args = append(args, driver.NamedValue{Name: "endTime", Value: *endTime})
	}

	query += fmt.Sprintf(" GROUP BY event_name, %s ORDER BY %s", groupCol, groupCol)

	rows, err := r.conn.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query metrics: %w", err)
	}
	defer rows.Close()

	var results []MetricRow
	for rows.Next() {
		var row MetricRow
		if err := rows.Scan(&row.EventName, &row.GroupKey, &row.TotalCount, &row.UniqueUsers); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		results = append(results, row)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return results, nil
}
