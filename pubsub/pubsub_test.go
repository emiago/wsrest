package pubsub

import (
	"encoding/json"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitPubSub(t *testing.T) {
	p := NewPubSub()

	assert.NotNil(t, p.recv)
	assert.NotNil(t, p.topics)
}

func TestSubscribtion(t *testing.T) {
	p := NewPubSub()

	s := Subscriber{
		Id:     "MyCustomId",
		Topics: []string{"go", "python"},
		Pipe:   make(chan []byte),
	}
	<-p.Subscribe(s) //Wait to be completed

	//CHeck is set subscribe
	gots, exists := p.subscribers[s.Id]
	assert.True(t, exists)
	assert.Equal(t, s.Topics, gots.Topics, "Subscribers not same")
	assert.Equal(t, s.Id, gots.Id, "Subscribers not same")
	assert.Equal(t, s.Pipe, gots.Pipe, "Subscribers not same")
}

func TestPublish(t *testing.T) {
	p := NewPubSub()
	logrus.SetLevel(logrus.DebugLevel)

	s := Subscriber{
		Id:     "MyCustomId",
		Topics: []string{"go", "python"},
		Pipe:   make(chan []byte, 3),
	}
	<-p.Subscribe(s) //Wait to be completed

	var e Eventer
	e = &Event{Type: "GoMsg", Topic: "go"}

	subs := p.findTopicSubscribers(e.GetTopic())
	require.Equal(t, len(subs), 1)
	require.Equal(t, subs[0].Pipe, s.Pipe)
	require.Equal(t, subs[0].Topics, s.Topics)
	//Test pipes
	t.Log("Testing pipes")
	subs[0].Pipe <- []byte("aaa")
	<-s.Pipe

	p.Publish(e)

	//CHeck is set subscribe
	rec := <-s.Pipe
	require.Greater(t, len(rec), 0)
	got := &Event{}
	json.Unmarshal(rec, got)
	assert.Equal(t, e, got, "Event got is not same")
}

func BenchmarkPublishingProcess(t *testing.B) {
	p := NewPubSub()
	s := Subscriber{
		Id:     "MyCustomId",
		Topics: []string{"go", "python"},
		Pipe:   make(chan []byte),
	}
	<-p.Subscribe(s) //Wait to be completed

	go func() {
		for {
			//Dumping messages
			_, more := <-s.Pipe
			if !more {
				break
			}
		}
	}()

	t.ResetTimer()
	t.StartTimer()
	for i := 0; i < t.N; i++ {
		p.Publish(&Event{Type: "GoMsg", Topic: "go"})
	}
	t.StopTimer()

	<-p.Unsubscribe(s)
	close(s.Pipe)
}
