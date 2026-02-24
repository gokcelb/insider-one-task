package metrics

import (
	"context"

	"github.com/insider/event-ingestion/clickhouse/repository"
)

type Service struct {
	repo *repository.MetricsRepository
}

func NewService(repo *repository.MetricsRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetMetrics(ctx context.Context, query MetricsQuery) ([]Metric, error) {
	rows, err := s.repo.GetMetrics(ctx, query.EventName, query.From, query.To, query.GroupBy)
	if err != nil {
		return nil, err
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
