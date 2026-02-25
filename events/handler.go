package events

import (
	"net/http"

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

type EventRequest struct {
	EventName  string         `json:"event_name" binding:"required"`
	Channel    string         `json:"channel" binding:"omitempty,oneof=web mobile api email push"`
	CampaignID string         `json:"campaign_id" binding:"omitempty"`
	UserID     string         `json:"user_id" binding:"required"`
	Timestamp  int64          `json:"timestamp" binding:"required"`
	Tags       []string       `json:"tags" binding:"omitempty"`
	Metadata   map[string]any `json:"metadata" binding:"omitempty"`
}

func (r *EventRequest) toEvent() Event {
	return Event{
		EventName:  r.EventName,
		Channel:    r.Channel,
		CampaignID: r.CampaignID,
		UserID:     r.UserID,
		Timestamp:  r.Timestamp,
		Tags:       r.Tags,
		Metadata:   r.Metadata,
	}
}

type BulkEventRequest struct {
	Events []EventRequest `json:"events" binding:"required,dive"`
}

type EventResponse struct {
	Status string `json:"status"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func (h *Handler) PostEvent(c *gin.Context) {
	var req EventRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	event := req.toEvent()

	if err := h.service.ProcessEvent(c.Request.Context(), event); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "internal server error",
		})
		return
	}

	c.JSON(http.StatusAccepted, EventResponse{
		Status: "accepted",
	})
}

func (h *Handler) PostEventBulk(c *gin.Context) {
	var req BulkEventRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: err.Error(),
		})
		return
	}

	events := make([]Event, len(req.Events))
	for i, r := range req.Events {
		events[i] = r.toEvent()
	}

	if err := h.service.ProcessBulk(c.Request.Context(), events); err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "internal server error",
		})
		return
	}

	c.JSON(http.StatusAccepted, EventResponse{
		Status: "accepted",
	})
}

func (h *Handler) RegisterRoutes(r *gin.Engine) {
	r.POST("/events", h.PostEvent)
	r.POST("/events/bulk", h.PostEventBulk)
}
