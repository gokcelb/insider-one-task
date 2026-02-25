package events

import (
	"context"
	"fmt"
	"strconv"

	"github.com/cespare/xxhash/v2"
	"github.com/insider/event-ingestion/kafka"
)

type eventPublisher interface {
	Publish(ctx context.Context, msg kafka.EventMessage) error
	PublishBulk(ctx context.Context, msgs []kafka.EventMessage) error
}

type Service struct {
	publisher eventPublisher
}

func NewService(publisher eventPublisher) *Service {
	return &Service{
		publisher: publisher,
	}
}

func (s *Service) ProcessEvent(ctx context.Context, event Event) error {
	event.EventHash = generateEventHash(event.EventName, event.UserID, event.Timestamp)
	msg, err := event.ToKafkaMessage()
	if err != nil {
		return fmt.Errorf("failed to convert event to kafka message: %w", err)
	}

	if err := s.publisher.Publish(ctx, msg); err != nil {
		return fmt.Errorf("failed to publish event: %w", err)
	}
	return nil
}

func (s *Service) ProcessBulk(ctx context.Context, events []Event) error {
	msgs := make([]kafka.EventMessage, len(events))
	for i := range events {
		events[i].EventHash = generateEventHash(events[i].EventName, events[i].UserID, events[i].Timestamp)
		msg, err := events[i].ToKafkaMessage()
		if err != nil {
			return fmt.Errorf("failed to convert event at index %d to kafka message: %w", i, err)
		}
		msgs[i] = msg
	}

	if err := s.publisher.PublishBulk(ctx, msgs); err != nil {
		return fmt.Errorf("failed to publish events: %w", err)
	}
	return nil
}

func generateEventHash(eventName, userID string, timestamp int64) uint64 {
	var buf [128]byte
	b := buf[:0]
	b = append(b, eventName...)
	b = append(b, ':')
	b = append(b, userID...)
	b = append(b, ':')
	b = strconv.AppendInt(b, timestamp, 10)
	return xxhash.Sum64(b)
}
