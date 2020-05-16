package kv

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKVFromName(t *testing.T) {
	kv, err := KVFromName("in-memory")
	assert.Nil(t, err)
	assert.Equal(t, INMEMORY, kv)

	kv2, err2 := KVFromName("foo")
	assert.Equal(t, Kind(""), kv2)
	assert.EqualError(t, err2, fmt.Sprintf("[%s] foo is not a supported implementation of KV interface", NotImplemented))
}
