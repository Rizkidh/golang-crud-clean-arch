// internal/kafka/producer.go
package kafka

import (
	"golang-crud-clean-arch/internal/event"
)

var kafkaEventPublisher event.EventPublisher // Global instance untuk publisher Kafka

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
