package messaging

type (
	Event interface {
		Topic() string
	}
)
