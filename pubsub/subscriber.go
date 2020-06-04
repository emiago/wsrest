package pubsub

type Subscriber struct {
	Id     string
	Topics []string
	Pipe   chan []byte
}
