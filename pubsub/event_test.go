package pubsub

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEvent(t *testing.T) {
	for _, e := range []EventType{ConfigCreated, ConfigUpdated, ConfigDeleted} {
		ev := NewEvent(e, "foo/config.yaml")
		assert.NotNil(t, ev)
		assert.Equal(t, fmt.Sprintf("[%s] - %s: %s", ev.CreatedAt(), ev.Kind(), ev.ConfigPath()), ev.String())
	}
}
