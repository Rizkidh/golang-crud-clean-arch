package event

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/segmentio/kafka-go"
)

type KafkaPublisher struct {
	writer *kafka.Writer
	topic  string
}

func NewKafkaPublisher(brokers []string, topic string) *KafkaPublisher {
	writer := &kafka.Writer{
		Addr:     kafka.TCP(brokers...),
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	}

	return &KafkaPublisher{
		writer: writer,
		topic:  topic,
	}
}

func (p *KafkaPublisher) Publish(ctx context.Context, topic string, key string, value interface{}) error {
	bytes, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	msg := kafka.Message{
		Key:   []byte(key),
		Value: bytes,
	}

	return p.writer.WriteMessages(ctx, msg)
}
