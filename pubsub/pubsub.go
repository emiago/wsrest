package pubsub

import (
	"encoding/json"

	"github.com/sirupsen/logrus"
)

type Marshaler interface {
	Marshal(v interface{}) ([]byte, error)
}

type JSONMarshaler struct{}

func (m *JSONMarshaler) Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

type Option interface{}
type OptionDoAsync struct{}

type eventSubAction struct {
	Event
	Subscriber *Subscriber
	Done       chan struct{}
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
		recv:        make(chan Eventer, 100),
		topics:      make(map[string][]*Subscriber),
		subscribers: make(map[string]*Subscriber),
		marshaler:   &JSONMarshaler{},
		log:         logrus.StandardLogger(),
	}

	go p.listen()
	return p
}

func (p *PubSub) SetMarshaler(m Marshaler) {
	p.marshaler = m
}

func (p *PubSub) Subscribe(s Subscriber) (done chan struct{}) {
	e := &eventSubAction{
		Event: Event{Type: "Subscribe"}, //
		Subscriber: &Subscriber{ //Always make a copy
			Id:     s.Id,
			Topics: append([]string{}, s.Topics...),
			Pipe:   s.Pipe,
		},
		Done: make(chan struct{}),
	}
	p.recv <- e
	return e.Done
}

func (p *PubSub) Unsubscribe(s Subscriber) (done chan struct{}) {
	e := &eventSubAction{
		Event: Event{Type: "Unsubscribe"}, //
		Subscriber: &Subscriber{ //We just need id and topics
			Id:     s.Id,
			Topics: append([]string{}, s.Topics...),
		},
		Done: make(chan struct{}),
	}
	p.recv <- e
	return e.Done
}

func (p *PubSub) Publish(e Eventer) {
	p.recv <- e
}

func (p *PubSub) listen() {
	for {
		select {
		case e := <-p.recv:
			p.processEvent(e)
		}
	}
}

func (p *PubSub) processEvent(e Eventer) {
	p.log.WithFields(logrus.Fields{
		"Type":  e.GetType(),
		"Topic": e.GetTopic(),
	}).Debug("Got event")

	if e.GetType() == "Subscribe" {
		m := e.(*eventSubAction)
		p.subscribe(m.Subscriber)
		close(m.Done)
		return
	}

	if e.GetType() == "Unsubscribe" {
		m := e.(*eventSubAction)
		p.unsubscribe(m.Subscriber)
		close(m.Done)
		return
	}

	p.handleEvent(e)
}

func (p *PubSub) handleEvent(e Eventer) {
	subs := p.findTopicSubscribers(e.GetTopic())
	if len(subs) == 0 {
		p.log.WithField("topic", e.GetTopic()).Debug("No subscribers for topic")
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
	for _, t := range s.Topics { //Always copy
		p.topics[t] = append(p.topics[t], s)
	}
	p.subscribers[s.Id] = s
}

func (p *PubSub) unsubscribe(s *Subscriber) {
	for _, t := range s.Topics {
		topicsubs := p.topics[t]
		for i, cs := range topicsubs {
			if s == cs || s.Id == cs.Id {
				topicsubs = append(topicsubs[:i], topicsubs[i+1:]...)
				break
			}
		}
	}
	delete(p.subscribers, s.Id)
}

func (p *PubSub) findTopicSubscribers(topic string) []*Subscriber {
	subs, exists := p.topics[topic]
	if !exists {
		return []*Subscriber{}
	}
	return subs
}

func (p *PubSub) broadcast(data []byte, sub []*Subscriber) {
	for _, s := range sub {
		s.Pipe <- data
	}
}
