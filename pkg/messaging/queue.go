package messaging

import (
	"context"
	"encoding/json"
)

type (
	QueueSender interface {
		Send(ctx context.Context, message QueueMessage) error
	}

	QueueMessage struct {
		EventName string
		Payload   any
		GroupID   *string
	}
)

func (q QueueMessage) Encode() ([]byte, error) {
	return json.Marshal(q.Payload)
}
