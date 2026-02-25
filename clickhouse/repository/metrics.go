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
	GroupKey    string
	TotalCount  uint64
	UniqueUsers uint64
}

func NewMetricsRepository(conn driver.Conn) *MetricsRepository {
	return &MetricsRepository{conn: conn}
}

func (r *MetricsRepository) GetMetrics(ctx context.Context, eventName string, startTime, endTime *time.Time, groupBy string) ([]MetricRow, error) {
	var groupCol string
	switch groupBy {
	case "channel":
		groupCol = "channel"
	case "hour":
		groupCol = "toString(toStartOfHour(timestamp))"
	case "day":
		groupCol = "toString(toStartOfDay(timestamp))"
	}

	selectClause := "count() AS total_count, uniq(user_id) AS unique_users"
	if groupCol != "" {
		selectClause = fmt.Sprintf("%s AS group_key, %s", groupCol, selectClause)
	}

	query := fmt.Sprintf("SELECT %s FROM events_db.events WHERE event_name = @eventName", selectClause)

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

	if groupCol != "" {
		query += fmt.Sprintf(" GROUP BY %s ORDER BY %s", groupCol, groupCol)
	}

	rows, err := r.conn.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query metrics: %w", err)
	}
	defer rows.Close()

	var results []MetricRow
	for rows.Next() {
		var row MetricRow
		if groupCol != "" {
			if err := rows.Scan(&row.GroupKey, &row.TotalCount, &row.UniqueUsers); err != nil {
				return nil, fmt.Errorf("failed to scan row: %w", err)
			}
		} else {
			if err := rows.Scan(&row.TotalCount, &row.UniqueUsers); err != nil {
				return nil, fmt.Errorf("failed to scan row: %w", err)
			}
		}
		results = append(results, row)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return results, nil
}
