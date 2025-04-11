package kafka

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
)

// KafkaConsumer represents a consumer listening to a specific topic.
type KafkaConsumer struct {
	Brokers []string
	Topic   string
	GroupID string
	Handler func(message kafka.Message)
}

// Start starts the Kafka consumer and processes messages using the provided handler.
func (kc *KafkaConsumer) Start(ctx context.Context) error {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:           kc.Brokers,
		GroupID:           kc.GroupID,
		Topic:             kc.Topic,
		MinBytes:          10e3, // 10KB
		MaxBytes:          10e6, // 10MB
		CommitInterval:    time.Second,
		HeartbeatInterval: 3 * time.Second,
	})
	defer r.Close()

	log.Printf("✅ Kafka consumer started for topic '%s' [group: %s]", kc.Topic, kc.GroupID)

	for {
		m, err := r.ReadMessage(ctx)
		if err != nil {
			log.Printf("❌ Error reading message: %v", err)
			return err
		}

		// Handle the message
		go kc.Handler(m)
	}
}

// Example handler usage
func PrintMessageHandler(msg kafka.Message) {
	fmt.Printf("📥 Received message on topic '%s': key=%s value=%s\n",
		msg.Topic, string(msg.Key), string(msg.Value))
}
