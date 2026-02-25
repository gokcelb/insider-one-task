package events

import (
	"context"
	"fmt"
	"hash/fnv"

	"github.com/insider/event-ingestion/kafka"
)

type Service struct {
	producer *kafka.Producer
}

func NewService(producer *kafka.Producer) *Service {
	return &Service{
		producer: producer,
	}
}

func (s *Service) ProcessEvent(ctx context.Context, event Event) error {
	event.EventHash = generateEventHash(event.EventName, event.UserID, event.Timestamp)
	msg, err := event.ToKafkaMessage()
	if err != nil {
		return err
	}

	return s.producer.Publish(ctx, msg)
}

func (s *Service) ProcessBulk(ctx context.Context, events []Event) error {
	msgs := make([]kafka.EventMessage, len(events))
	for i := range events {
		events[i].EventHash = generateEventHash(events[i].EventName, events[i].UserID, events[i].Timestamp)
		msg, err := events[i].ToKafkaMessage()
		if err != nil {
			return err
		}
		msgs[i] = msg
	}

	return s.producer.PublishBulk(ctx, msgs)
}

func generateEventHash(eventName, userID string, timestamp int64) uint64 {
	hash := fnv.New64a()
	hash.Write([]byte(fmt.Sprintf("%s:%s:%d", eventName, userID, timestamp)))
	return hash.Sum64()
}
