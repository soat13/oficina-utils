package messaging

type Broker interface {
	Consumer
	Publisher
}
