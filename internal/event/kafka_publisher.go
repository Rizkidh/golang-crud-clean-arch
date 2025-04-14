package event

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/segmentio/kafka-go"
)

// KafkaPublisher adalah struct untuk publish event ke Kafka.
// Menyimpan writer per topik agar efisien dan thread-safe.
type KafkaPublisher struct {
	brokers []string                 // Daftar alamat broker Kafka
	writers map[string]*kafka.Writer // Map writer Kafka berdasarkan topik
	mu      sync.Mutex               // Mutex untuk menghindari race condition
}

// NewKafkaPublisher menginisialisasi KafkaPublisher dan membuat topik jika belum ada.
func NewKafkaPublisher(brokers []string, topic string) *KafkaPublisher {
	// Buat topik jika belum ada
	err := createKafkaTopic(brokers[0], topic)
	if err != nil {
		fmt.Printf("❌ Error creating topic %s: %v\n", topic, err)
	} else {
		fmt.Printf("✅ Topic %s created successfully\n", topic)
	}

	return &KafkaPublisher{
		brokers: brokers,
		writers: map[string]*kafka.Writer{
			topic: &kafka.Writer{
				Addr:     kafka.TCP(brokers...), // Alamat broker Kafka
				Topic:    topic,                 // Nama topik
				Balancer: &kafka.LeastBytes{},   // Load balancing berdasarkan ukuran pesan
			},
		},
	}
}

// getWriter mengembalikan writer untuk topik tertentu, atau membuat yang baru jika belum ada.
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

// Publish mengirimkan pesan ke Kafka dengan topik, key, dan value.
// Value akan di-encode menjadi JSON sebelum dikirim.
func (p *KafkaPublisher) Publish(ctx context.Context, topic string, key string, value interface{}) error {
	writer := p.getWriter(topic)

	// Konversi value ke format JSON
	bytes, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	// Buat pesan Kafka
	msg := kafka.Message{
		Key:   []byte(key),
		Value: bytes,
		Time:  time.Now(),
	}

	// Kirim pesan ke Kafka
	return writer.WriteMessages(ctx, msg)
}

// Close menutup semua writer Kafka untuk membebaskan resource.
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

// createKafkaTopic membuat topik Kafka jika belum ada di broker.
func createKafkaTopic(broker, topic string) error {
	// Koneksi awal ke broker Kafka
	conn, err := kafka.Dial("tcp", broker)
	if err != nil {
		return err
	}
	defer conn.Close()

	// Dapatkan informasi controller Kafka
	controller, err := conn.Controller()
	if err != nil {
		return err
	}

	// Koneksi langsung ke controller Kafka
	ctrlConn, err := kafka.Dial("tcp", fmt.Sprintf("%s:%d", controller.Host, controller.Port))
	if err != nil {
		return err
	}
	defer ctrlConn.Close()

	// Coba buat topik
	return ctrlConn.CreateTopics(kafka.TopicConfig{
		Topic:             topic,
		NumPartitions:     1,
		ReplicationFactor: 1,
	})
}
