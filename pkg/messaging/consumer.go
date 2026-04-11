package messaging

import (
	"context"
)

type (
	Handler func(ctx context.Context, msg Message) error

	Consumer interface {
		Subscribe(topic string, handler Handler)
		Listen(ctx context.Context)
		Stop()
	}
)
