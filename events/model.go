package events

import (
	"encoding/json"

	"github.com/insider/event-ingestion/kafka"
)

type Event struct {
	EventHash  uint64
	EventName  string
	Channel    string
	CampaignID string
	UserID     string
	Timestamp  int64
	Tags       []string
	Metadata   map[string]any
}

func (e *Event) ToKafkaMessage() kafka.EventMessage {
	metadataJSON, _ := json.Marshal(e.Metadata)

	return kafka.EventMessage{
		EventHash:  e.EventHash,
		EventName:  e.EventName,
		Channel:    e.Channel,
		CampaignID: e.CampaignID,
		UserID:     e.UserID,
		Timestamp:  e.Timestamp,
		Tags:       e.Tags,
		Metadata:   string(metadataJSON),
	}
}
