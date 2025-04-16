package event

import (
	"context"
	"encoding/json"
	"fmt"
)

// EventPublisher adalah interface untuk mempublikasikan event ke Kafka.
type EventPublisher interface {
	Publish(ctx context.Context, topic string, key string, value interface{}) error
}

func MarshalData(data interface{}) ([]byte, error) {
	marshaledData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data: %w", err)
	}
	return marshaledData, nil
}
