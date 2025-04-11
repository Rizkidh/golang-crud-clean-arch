package event

import (
	"context"
)

type EventPublisher interface {
	Publish(ctx context.Context, topic string, key string, value interface{}) error
}
