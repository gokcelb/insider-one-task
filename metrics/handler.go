package metrics

import (
	"log"
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
	GroupBy   string `form:"group_by" binding:"omitempty,oneof=channel hour"`
}

func (p *MetricsQueryParams) toMetricsQuery() MetricsQuery {
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

	return query
}

type MetricsResponse struct {
	EventName   string           `json:"event_name"`
	From        int64            `json:"from,omitempty"`
	To          int64            `json:"to,omitempty"`
	TotalEvents *uint64          `json:"total_events,omitempty"`
	UniqueUsers *uint64          `json:"unique_users,omitempty"`
	GroupedBy   string           `json:"grouped_by,omitempty"`
	Data        []MetricResponse `json:"data,omitempty"`
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
	resp := MetricsResponse{
		EventName: query.EventName,
	}

	if query.From != nil {
		resp.From = query.From.Unix()
	}
	if query.To != nil {
		resp.To = query.To.Unix()
	}

	if query.GroupBy == "" {
		var total, unique uint64
		if len(metrics) > 0 {
			total = metrics[0].TotalEvents
			unique = metrics[0].UniqueUsers
		}
		resp.TotalEvents = &total
		resp.UniqueUsers = &unique
	} else {
		resp.GroupedBy = query.GroupBy
		resp.Data = make([]MetricResponse, len(metrics))
		for i, m := range metrics {
			resp.Data[i] = MetricResponse{
				Group:       m.Group,
				TotalEvents: m.TotalEvents,
				UniqueUsers: m.UniqueUsers,
			}
		}
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

	query := params.toMetricsQuery()

	metrics, err := h.service.GetMetrics(c.Request.Context(), query)
	if err != nil {
		log.Printf("failed to fetch metrics: %v", err)
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
