package pubsub

import "github.com/sirupsen/logrus"

type Subscriber struct {
	pipe   chan []byte
	id     string
	topics []string
}

type Marshaler interface {
	Marshal(v interface{}) ([]byte, error)
}

type eventSubAction struct {
	Event
	Subscriber *Subscriber
}

type PubSub struct {
	subscribers map[string]*Subscriber
	topics      map[string][]*Subscriber
	recv        chan Eventer
	marshaler   Marshaler
	log         logrus.FieldLogger
}

func NewPubSub() *PubSub {
	p := &PubSub{
		recv:   make(chan Eventer, 100),
		topics: make(map[string][]*Subscriber),
	}

	go p.listen()
	return p
}

func (p *PubSub) SetMarshaler(m Marshaler) {
	p.marshaler = m
}

func (p *PubSub) Subscribe(s *Subscriber) {
	e := &eventSubAction{
		Event:      Event{Type: "Subscribe"}, //
		Subscriber: s,
	}
	p.recv <- e
}

func (p *PubSub) Unsubscribe(s *Subscriber) {
	e := &eventSubAction{
		Event:      Event{Type: "Unsubscribe"}, //
		Subscriber: s,
	}
	p.recv <- e
}

func (p *PubSub) listen() {
	select {
	case e := <-p.recv:
		p.processEvent(e)
	}
}

func (p *PubSub) processEvent(e Eventer) {
	if e.GetType() == "Subscribe" {
		m := e.(*eventSubAction)
		p.subscribe(m.Subscriber)
	}

	subs := p.findTopicSubscribers(e.GetTopic())
	if len(subs) == 0 {
		return
	}

	data, err := p.marshaler.Marshal(e)
	if err != nil {
		p.log.WithError(err).Error("Fail to marshal event")
		return
	}

	p.broadcast(data, subs)
}

func (p *PubSub) subscribe(s *Subscriber) {
	for _, t := range s.topics {
		p.topics[t] = append(p.topics[t], s)
	}
	p.subscribers[s.id] = s
}

func (p *PubSub) unsubscribe(s *Subscriber) {
	for _, t := range s.topics {
		topicsubs := p.topics[t]
		for i, cs := range topicsubs {
			if s == cs || s.id == cs.id {
				topicsubs = append(topicsubs[:i], topicsubs[i+1:]...)
				break
			}
		}
	}
	delete(p.subscribers, s.id)
}

func (p *PubSub) findTopicSubscribers(topic string) []*Subscriber {
	subs, exists := p.topics[topic]
	if !exists {
		return []*Subscriber{}
	}
	return subs
}

func (p *PubSub) broadcast(frame []byte, sub []*Subscriber) {
	for _, s := range sub {
		s.pipe <- frame
	}
}
