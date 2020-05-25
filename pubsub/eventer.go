package ws

import "time"

type Eventer interface {
	GetType() string
	SetType(t string)
	GetApplication() string
	SetApplication(t string)
	GetSource() string
	SetSource(t string)
	GetSourceId() string
	SetSourceId(t string)
}

type Event struct {
	Type        string `json:"type"`
	Application string `json:"application"`
	Source      string `json:"source"`
	SourceId    string `json:"source_id"`
	Timestamp   *Time  `json:"timestamp"`
}

func NewEvent(t string, source string, sourceid string, app string) Event {
	time := Time(time.Now())

	return Event{
		Type:        t,
		Application: app,
		Source:      source,
		SourceId:    sourceid,
		Timestamp:   &time,
	}
}

func (p *Event) GetType() string {
	return p.Type
}

func (p *Event) SetType(t string) {
	p.Type = t
}

func (p *Event) GetApplication() string {
	return p.Application
}

func (p *Event) SetApplication(t string) {
	p.Application = t
}

func (p *Event) GetSource() string {
	return p.Source
}

func (p *Event) SetSource(t string) {
	p.Source = t
}

func (p *Event) GetSourceId() string {
	return p.SourceId
}

func (p *Event) SetSourceId(t string) {
	p.SourceId = t
}
