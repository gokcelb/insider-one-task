package metrics

import (
	"context"
	"fmt"
	"time"

	"github.com/insider/event-ingestion/clickhouse/repository"
)

type metricsRepository interface {
	GetMetrics(ctx context.Context, eventName string, startTime, endTime *time.Time, groupBy string) ([]repository.MetricRow, error)
}

type Service struct {
	repo metricsRepository
}

func NewService(repo metricsRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetMetrics(ctx context.Context, query MetricsQuery) ([]Metric, error) {
	rows, err := s.repo.GetMetrics(ctx, query.EventName, query.From, query.To, query.GroupBy)
	if err != nil {
		return nil, fmt.Errorf("failed to get metrics: %w", err)
	}

	metrics := make([]Metric, len(rows))
	for i, row := range rows {
		metrics[i] = Metric{
			Group:       row.GroupKey,
			TotalEvents: row.TotalCount,
			UniqueUsers: row.UniqueUsers,
		}
	}

	return metrics, nil
}
