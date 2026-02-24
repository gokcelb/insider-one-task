package metrics

import "time"

type MetricsQuery struct {
	EventName string
	From      *time.Time
	To        *time.Time
	GroupBy   string
}

type Metric struct {
	Group       string
	TotalEvents uint64
	UniqueUsers uint64
}
