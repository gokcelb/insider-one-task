package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/twmb/franz-go/pkg/kgo"

	"github.com/insider/event-ingestion/config"
)

type Producer struct {
	client *kgo.Client
	topic  string
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
	opts := []kgo.Opt{
		kgo.SeedBrokers(cfg.Brokers...),
		kgo.DefaultProduceTopic(cfg.Topic),
		kgo.ProducerBatchCompression(kgo.Lz4Compression()),
		kgo.AllowAutoTopicCreation(),
	}

	client, err := kgo.NewClient(opts...)
	if err != nil {
		return nil, err
	}

	return &Producer{
		client: client,
		topic:  cfg.Topic,
	}, nil
}

func (p *Producer) Publish(ctx context.Context, msg EventMessage) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	record := &kgo.Record{
		Topic: p.topic,
		Key:   []byte(fmt.Sprintf("%d", msg.EventHash)),
		Value: data,
	}

	p.client.Produce(ctx, record, func(r *kgo.Record, err error) {
		if err != nil {
			log.Printf("async produce error: %v", err)
		}
	})

	return nil
}

func (p *Producer) PublishBulk(ctx context.Context, msgs []EventMessage) error {
	for _, msg := range msgs {
		if err := p.Publish(ctx, msg); err != nil {
			return err
		}
	}
	return nil
}

func (p *Producer) Close() {
	p.client.Close()
}

func (p *Producer) Ping(ctx context.Context) error {
	return p.client.Ping(ctx)
}
