package messaging

type QueueBroker interface {
	Consumer
	QueueSender
}
