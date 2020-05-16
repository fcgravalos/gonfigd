package kv

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInMemoryOperations(t *testing.T) {
	db, err1 := NewKV(INMEMORY)
	assert.NotNil(t, db)

	_, ok := db.(*InMemory)
	assert.True(t, ok)

	assert.Nil(t, err1)

	v1, err2 := db.Get("foo")
	assert.Nil(t, v1)
	assert.EqualError(t, err2, fmt.Sprintf("[%s] Key foo not found in KV", KeyNotFound))

	v2, err3 := NewValue([]byte("bar"))
	assert.Nil(t, err3)

	err4 := db.Put("foo", v2)
	assert.Nil(t, err4)

	v3, err5 := db.Get("foo")
	assert.Nil(t, err5)
	assert.Equal(t, v3, v2)

	err6 := db.Delete("foo")
	assert.Nil(t, err6)

	v4, err7 := db.Get("foo")
	assert.Nil(t, v4)
	assert.EqualError(t, err7, fmt.Sprintf("[%s] Key foo not found in KV", KeyNotFound))
}
