package event

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/segmentio/kafka-go"
)

type KafkaPublisher struct {
	brokers []string
	writers map[string]*kafka.Writer
	mu      sync.Mutex
}

// NewKafkaPublisher initializes the publisher and creates the topic if not exists
func NewKafkaPublisher(brokers []string, topic string) *KafkaPublisher {
	// Check and create topic
	err := createKafkaTopic(brokers[0], topic)
	if err != nil {
		fmt.Printf("⚠️ Failed to create Kafka topic '%s': %v\n", topic, err)
	}

	return &KafkaPublisher{
		brokers: brokers,
		writers: map[string]*kafka.Writer{
			topic: &kafka.Writer{
				Addr:     kafka.TCP(brokers...),
				Topic:    topic,
				Balancer: &kafka.LeastBytes{},
			},
		},
	}
}

func (p *KafkaPublisher) getWriter(topic string) *kafka.Writer {
	p.mu.Lock()
	defer p.mu.Unlock()

	writer, exists := p.writers[topic]
	if !exists {
		writer = &kafka.Writer{
			Addr:     kafka.TCP(p.brokers...),
			Topic:    topic,
			Balancer: &kafka.LeastBytes{},
		}
		p.writers[topic] = writer
	}
	return writer
}

func (p *KafkaPublisher) Publish(ctx context.Context, topic string, key string, value interface{}) error {
	writer := p.getWriter(topic)

	bytes, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	msg := kafka.Message{
		Key:   []byte(key),
		Value: bytes,
		Time:  time.Now(),
	}

	return writer.WriteMessages(ctx, msg)
}

func (p *KafkaPublisher) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	var firstErr error
	for topic, writer := range p.writers {
		if err := writer.Close(); err != nil && firstErr == nil {
			firstErr = fmt.Errorf("error closing writer for topic %s: %w", topic, err)
		}
	}
	return firstErr
}

// createKafkaTopic creates a Kafka topic if it doesn't exist
func createKafkaTopic(broker, topic string) error {
	conn, err := kafka.Dial("tcp", broker)
	if err != nil {
		return err
	}
	defer conn.Close()

	controller, err := conn.Controller()
	if err != nil {
		return err
	}

	ctrlConn, err := kafka.Dial("tcp", fmt.Sprintf("%s:%d", controller.Host, controller.Port))
	if err != nil {
		return err
	}
	defer ctrlConn.Close()

	// Try to create topic
	return ctrlConn.CreateTopics(kafka.TopicConfig{
		Topic:             topic,
		NumPartitions:     1,
		ReplicationFactor: 1,
	})
}
