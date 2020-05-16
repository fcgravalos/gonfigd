package kv

import (
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetValue(t *testing.T) {
	v, err := NewValue([]byte("foo"))
	assert.Nil(t, err)
	assert.Equal(t, "foo", v.Text())
	assert.Equal(t, fmt.Sprintf("%x", md5.Sum([]byte("foo"))), v.MD5())
	assert.NotNil(t, v.LastModified())
	_, err2 := base64.StdEncoding.DecodeString(v.Data())
	assert.Nil(t, err2)
}
