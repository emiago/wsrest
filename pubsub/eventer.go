package pubsub

import "time"

type Eventer interface {
	GetType() string
	SetType(t string)
	GetTopic() string
	SetTopic(t string)
	GetTopicId() string
	SetTopicId(t string)
}

type Event struct {
	Type        string `json:"type"`
	Application string `json:"application"`
	Topic       string `json:"topic"`
	TopicId     string `json:"topic_id"`
	Timestamp   *Time  `json:"timestamp"`
}

func NewEvent(t string, Topic string, Topicid string) Event {
	time := Time(time.Now())

	return Event{
		Type:      t,
		Topic:     Topic,
		TopicId:   Topicid,
		Timestamp: &time,
	}
}

func (p *Event) GetType() string {
	return p.Type
}

func (p *Event) SetType(t string) {
	p.Type = t
}

func (p *Event) GetTopic() string {
	return p.Topic
}

func (p *Event) SetTopic(t string) {
	p.Topic = t
}

func (p *Event) GetTopicId() string {
	return p.TopicId
}

func (p *Event) SetTopicId(t string) {
	p.TopicId = t
}
