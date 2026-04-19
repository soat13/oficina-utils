package messaging

import "encoding/json"

type Message struct {
	Payload []byte
}

func DecodePayload[T any](msg Message) (*T, error) {
	var payload T

	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		return nil, err
	}

	return &payload, nil
}
