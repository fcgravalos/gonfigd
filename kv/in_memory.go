package kv

import (
	"sync"
)

// InMemory is an in-memory data structure implementation of the KV interface
type InMemory struct {
	sync.Mutex
	Db map[string]*Value
}

// Put inserts a new key/Value pair in the KV Db
// It wont raise any error as the operation is safe
func (im *InMemory) Put(key string, value *Value) error {
	im.Lock()
	im.Db[key] = value
	im.Unlock()
	return nil
}

// Get retrieves the value of the given key
func (im *InMemory) Get(key string) (*Value, error) {
	v, ok := im.Db[key]
	if !ok {
		return nil, NewKeyNotFoundError(key)
	}
	return v, nil
}

// Delete will remove a key from the KV Db
func (im *InMemory) Delete(key string) error {
	im.Lock()
	delete(im.Db, key)
	im.Unlock()
	return nil
}
