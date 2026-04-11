package messaging

import (
	"context"
	"encoding/json"
)

type (
	Publisher interface {
		Publish(ctx context.Context, topic string, payload []byte) error
	}

	Event interface {
		Topic() string
	}
)

func Publish(ctx context.Context, pub Publisher, event Event) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return pub.Publish(ctx, event.Topic(), payload)
}
