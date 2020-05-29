package pubsub

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInitPubSub(t *testing.T) {
	p := NewPubSub()

	assert.NotNil(t, p.recv)
	assert.NotNil(t, p.topics)
}
