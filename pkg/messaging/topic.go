package messaging

import (
	"context"
	"encoding/json"
)

type (
	TopicPublisher interface {
		Publish(ctx context.Context, message TopicMessage) error
	}

	TopicMessage struct {
		EventName string
		Payload   any
		GroupID   *string
	}
)

func (q TopicMessage) Encode() (string, error) {
	encoded, err := json.Marshal(q.Payload)
	if err != nil {
		return "", err
	}

	return string(encoded), nil
}
