// internal/kafka/consumer.go
package kafka

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
)

// KafkaConsumer merupakan struktur yang menangani konsumsi pesan dari Kafka.
type KafkaConsumer struct {
	Brokers []string                    // Daftar broker Kafka
	Topic   string                      // Nama topik Kafka
	GroupID string                      // ID grup consumer
	Handler func(message kafka.Message) // Fungsi handler untuk memproses pesan yang diterima
}

// Start menjalankan Kafka consumer dan membaca pesan dari topik secara kontinu.
func (kc *KafkaConsumer) Start(ctx context.Context) error {
	// Konfigurasi reader Kafka
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:           kc.Brokers,      // Daftar broker
		GroupID:           kc.GroupID,      // ID group (untuk shared consumer)
		Topic:             kc.Topic,        // Nama topik yang dikonsumsi
		MinBytes:          10e3,            // Minimum ukuran pesan (10KB)
		MaxBytes:          10e6,            // Maksimum ukuran pesan (10MB)
		CommitInterval:    time.Second,     // Interval commit offset
		HeartbeatInterval: 3 * time.Second, // Interval heartbeat untuk Kafka group
	})
	defer r.Close()

	log.Printf("✅ Kafka consumer started for topic '%s' [group: %s]", kc.Topic, kc.GroupID)

	for {
		// Baca pesan dari Kafka
		m, err := r.ReadMessage(ctx)
		if err != nil {
			log.Printf("❌ Error reading message: %v", err)
			return err
		}

		// Jalankan handler untuk memproses pesan secara async
		go kc.Handler(m)
	}
}

// PrintMessageHandler adalah contoh handler default untuk mencetak isi pesan ke konsol.
func PrintMessageHandler(msg kafka.Message) {
	fmt.Printf("📥 Received message on topic '%s':\n🔑 Key: %s\n📦 Value: %s\n\n",
		msg.Topic, string(msg.Key), string(msg.Value))
}
