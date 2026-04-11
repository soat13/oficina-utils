package messaging

import (
	"context"
)

type (
	Handler func(ctx context.Context, topic string, payload []byte) error

	Consumer interface {
		Subscribe(topic string, handler Handler)
		Listen(ctx context.Context)
		Stop()
	}
)
