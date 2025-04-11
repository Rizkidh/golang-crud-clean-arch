// internal/kafka/producer.go
package kafka

import (
	"context"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
)

var kafkaWriter *kafka.Writer

func InitKafkaWriter(broker, topic string) {
	kafkaWriter = &kafka.Writer{
		Addr:         kafka.TCP(broker),
		Topic:        topic,
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: kafka.RequireAll,
	}
	log.Println("Kafka writer initialized")
}

func PublishMessage(ctx context.Context, key, message string) error {
	if kafkaWriter == nil {
		return nil
	}
	msg := kafka.Message{
		Key:   []byte(key),
		Value: []byte(message),
		Time:  time.Now(),
	}
	log.Printf("Publishing message to Kafka: %s\n", msg.Value)
	return kafkaWriter.WriteMessages(ctx, msg)
}
