package pubsub

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInMemoryOperations(t *testing.T) {
	ps, err1 := NewPubSub(INMEMORY)
	assert.NotNil(t, ps)

	_, ok := ps.(*InMemory)
	assert.True(t, ok)

	assert.Nil(t, err1)

	err2 := ps.CreateTopic("foo")
	assert.Nil(t, err2)
	assert.True(t, ps.TopicExists("foo"))

	s, err3 := ps.Subscribe("bar")
	assert.Nil(t, s)
	assert.EqualError(t, err3, fmt.Sprintf("[%s] Topic bar does not exist", NoSuchTopic))

	s2, err4 := ps.Subscribe("foo")
	assert.NotNil(t, s2)
	assert.Nil(t, err4)

	go func(s *Subscription) {
		ch := s.Channel()
		<-ch
	}(s2)

	err5 := ps.Publish("foo", NewEvent(ConfigCreated, "foo"))
	assert.Nil(t, err5)

	err6 := ps.UnSubscribe("foo", s2.ID())
	assert.Nil(t, err6)

	err7 := ps.DeleteTopic("foo")
	assert.Nil(t, err7)

	err8 := ps.UnSubscribe("foo", s2.ID())
	assert.EqualError(t, err8, fmt.Sprintf("[%s] Topic foo does not exist", NoSuchTopic))
}
