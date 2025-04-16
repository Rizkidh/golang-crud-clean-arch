package kafka

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"golang-crud-clean-arch/internal/event"
	"log"

	"github.com/segmentio/kafka-go"
)

// KafkaEventPublisher adalah instance global untuk publisher Kafka.
var kafkaEventPublisher event.EventPublisher

// InitKafkaPublisher menginisialisasi Kafka publisher dengan broker dan topik default.
// Fungsi ini bisa dipanggil saat aplikasi start (misalnya di main.go).
func InitKafkaPublisher(brokers []string, topic string) {
	kafkaEventPublisher = event.NewKafkaPublisher(brokers, topic)
}

// GetKafkaPublisher mengembalikan instance global dari Kafka publisher.
// Digunakan oleh komponen lain (misalnya usecase) untuk publish event.
func GetKafkaPublisher() event.EventPublisher {
	return kafkaEventPublisher
}

// Compress menggunakan GZIP untuk mengompresi data.
func Compress(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	gzipWriter := gzip.NewWriter(&buf)
	_, err := gzipWriter.Write(data)
	if err != nil {
		return nil, fmt.Errorf("failed to compress data: %w", err)
	}
	if err := gzipWriter.Close(); err != nil {
		return nil, fmt.Errorf("failed to close gzip writer: %w", err)
	}
	return buf.Bytes(), nil
}

// PublishEvent mengirimkan event yang sudah terkompresi ke Kafka menggunakan kafka-go.
func PublishEvent(topic string, eventType string, data interface{}, brokers []string) error {
	// Convert event data menjadi byte slice
	eventData, err := event.MarshalData(data)
	if err != nil {
		return fmt.Errorf("failed to marshal event data: %w", err)
	}

	// Kompresi data menggunakan GZIP
	compressedData, err := Compress(eventData)
	if err != nil {
		return fmt.Errorf("failed to compress event data: %w", err)
	}

	// Membuat koneksi Kafka menggunakan kafka-go
	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers: brokers,
		Topic:   topic,
	})

	// Kirim pesan ke Kafka
	err = writer.WriteMessages(
		context.Background(),
		kafka.Message{
			Value: compressedData, // Mengirimkan data terkompresi
		},
	)
	if err != nil {
		return fmt.Errorf("failed to send message to Kafka: %w", err)
	}

	log.Printf("Event of type %s published to Kafka on topic %s", eventType, topic)
	return nil
}
