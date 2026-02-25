package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	kafkago "github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/compress"

	"github.com/insider/event-ingestion/config"
)

type Producer struct {
	writer *kafkago.Writer
	addr   string
}

type EventMessage struct {
	EventHash  uint64   `json:"event_hash"`
	EventName  string   `json:"event_name"`
	Channel    string   `json:"channel"`
	CampaignID string   `json:"campaign_id"`
	UserID     string   `json:"user_id"`
	Timestamp  int64    `json:"timestamp"`
	Tags       []string `json:"tags"`
	Metadata   string   `json:"metadata"`
}

func NewProducer(cfg config.KafkaConfig) (*Producer, error) {
	writer := &kafkago.Writer{
		Addr:         kafkago.TCP(cfg.Brokers...),
		Topic:        cfg.Topic,
		Balancer:     &kafkago.LeastBytes{},
		Compression:  compress.Lz4,
		RequiredAcks: kafkago.RequireAll,
	}

	return &Producer{
		writer: writer,
		addr:   cfg.Brokers[0],
	}, nil
}

func (p *Producer) Publish(ctx context.Context, msg EventMessage) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	return p.writer.WriteMessages(ctx, kafkago.Message{
		Key:   strconv.AppendUint(nil, msg.EventHash, 10),
		Value: data,
	})
}

func (p *Producer) PublishBulk(ctx context.Context, msgs []EventMessage) error {
	messages := make([]kafkago.Message, len(msgs))
	for i, msg := range msgs {
		data, err := json.Marshal(msg)
		if err != nil {
			return err
		}
		messages[i] = kafkago.Message{
			Key:   strconv.AppendUint(nil, msg.EventHash, 10),
			Value: data,
		}
	}

	return p.writer.WriteMessages(ctx, messages...)
}

func (p *Producer) Close() {
	p.writer.Close()
}

func (p *Producer) Ping(ctx context.Context) error {
	conn, err := kafkago.Dial("tcp", p.addr)
	if err != nil {
		return fmt.Errorf("failed to connect to kafka: %w", err)
	}
	defer conn.Close()

	if deadline, ok := ctx.Deadline(); ok {
		conn.SetDeadline(deadline)
	}

	_, err = conn.Brokers()
	if err != nil {
		return fmt.Errorf("failed to list brokers: %w", err)
	}

	return nil
}
