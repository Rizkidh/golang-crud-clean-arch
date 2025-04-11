package kafka

import (
	"context"
	"log"

	"github.com/segmentio/kafka-go"
)

func StartKafkaConsumer(broker, topic, groupID string, handler func(string)) {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{broker},
		Topic:    topic,
		GroupID:  groupID,
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
	})

	log.Printf("Kafka consumer listening on topic: %s", topic)

	for {
		m, err := r.ReadMessage(context.Background())
		if err != nil {
			log.Printf("Error reading Kafka message: %v", err)
			continue
		}
		handler(string(m.Value))
	}
}
