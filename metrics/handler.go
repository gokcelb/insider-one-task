package metrics

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

type MetricsQueryParams struct {
	EventName string `form:"event_name" binding:"required"`
	From      int64  `form:"from" binding:"omitempty"`
	To        int64  `form:"to" binding:"omitempty"`
	GroupBy   string `form:"group_by" binding:"omitempty"`
}

func (p *MetricsQueryParams) toMetricsQuery() (MetricsQuery, error) {
	query := MetricsQuery{
		EventName: p.EventName,
		GroupBy:   p.GroupBy,
	}

	if p.From > 0 {
		t := time.Unix(p.From, 0).UTC()
		query.From = &t
	}

	if p.To > 0 {
		t := time.Unix(p.To, 0).UTC()
		query.To = &t
	}

	return query, nil
}

type MetricsResponse struct {
	EventName   string           `json:"event_name"`
	From        int64            `json:"from,omitempty"`
	To          int64            `json:"to,omitempty"`
	GroupedBy   string           `json:"grouped_by"`
	TotalEvents uint64           `json:"total_events"`
	UniqueUsers uint64           `json:"unique_users"`
	Data        []MetricResponse `json:"data"`
}

type MetricResponse struct {
	Group       string `json:"group"`
	TotalEvents uint64 `json:"total_events"`
	UniqueUsers uint64 `json:"unique_users"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func toMetricsResponse(query MetricsQuery, metrics []Metric) MetricsResponse {
	data := make([]MetricResponse, len(metrics))
	var totalEvents, uniqueUsers uint64

	for i, m := range metrics {
		data[i] = MetricResponse{
			Group:       m.Group,
			TotalEvents: m.TotalEvents,
			UniqueUsers: m.UniqueUsers,
		}
		totalEvents += m.TotalEvents
		uniqueUsers += m.UniqueUsers
	}

	resp := MetricsResponse{
		EventName:   query.EventName,
		GroupedBy:   query.GroupBy,
		TotalEvents: totalEvents,
		UniqueUsers: uniqueUsers,
		Data:        data,
	}

	if query.From != nil {
		resp.From = query.From.Unix()
	}
	if query.To != nil {
		resp.To = query.To.Unix()
	}

	return resp
}

func (h *Handler) GetMetrics(c *gin.Context) {
	var params MetricsQueryParams

	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	query, err := params.toMetricsQuery()
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "invalid time format, use RFC3339",
		})
		return
	}

	metrics, err := h.service.GetMetrics(c.Request.Context(), query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "failed to fetch metrics",
		})
		return
	}

	c.JSON(http.StatusOK, toMetricsResponse(query, metrics))
}

func (h *Handler) RegisterRoutes(r *gin.Engine) {
	r.GET("/metrics", h.GetMetrics)
}
