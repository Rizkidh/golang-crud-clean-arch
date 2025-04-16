package kafka

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/segmentio/kafka-go"
)

type KafkaConsumer struct {
	Brokers []string
	Topic   string
	GroupID string
	Handler func(message kafka.Message)
}

func (kc *KafkaConsumer) Start(ctx context.Context) error {
	// Create Kafka reader
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        kc.Brokers,
		GroupID:        kc.GroupID,
		Topic:          kc.Topic,
		MinBytes:       10e3,
		MaxBytes:       10e6,
		CommitInterval: 0,
	})
	defer r.Close()

	log.Printf("📥 Kafka Consumer started for topic '%s'", kc.Topic)

	// Signal channel for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start consumer loop
	for {
		select {
		case <-sigChan:
			log.Println("Graceful shutdown initiated")
			return nil
		default:
			m, err := r.ReadMessage(ctx)
			if err != nil {
				log.Printf("❌ Error reading message from topic '%s': %v", kc.Topic, err)
				return err
			}

			kc.Handler(m)
			log.Printf("✅ Message processed from topic '%s'", kc.Topic)
		}
	}
}
