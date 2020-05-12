package pubsub

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKVFromName(t *testing.T) {
	ps, err := PubSubFromName("in-memory")
	assert.Nil(t, err)
	assert.Equal(t, INMEMORY, ps)

	ps2, err2 := PubSubFromName("foo")
	assert.Equal(t, Kind(""), ps2)
	assert.EqualError(t, err2, fmt.Sprintf("[%s] foo is not a supported implementation of PubSub interface", NotImplemented))
}
