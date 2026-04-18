package messaging

import (
	"context"
	"encoding/json"
)

type (
	Publisher interface {
		Publish(ctx context.Context, topic string, payload []byte, opts ...PublishOption) error
	}

	Event interface {
		Topic() string
	}

	PublishOptions struct {
		MessageGroupId string
	}

	PublishOption func(*PublishOptions)
)

func Publish(ctx context.Context, pub Publisher, event Event, opts ...PublishOption) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return pub.Publish(ctx, event.Topic(), payload, opts...)
}
