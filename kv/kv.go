package kv

import (
	"fmt"
)

const (
	INMEMORY Kind = "in-memory"
)

var supportedKVs map[Kind]struct{}

type KV interface {
	Put(k string, v *Value) error
	Get(k string) (*Value, error)
	Delete(k string) error
}

type Kind string

func init() {
	supportedKVs = map[Kind]struct{}{INMEMORY: struct{}{}}
}

func NewKV(kind Kind) (KV, error) {
	var kv KV
	if _, ok := supportedKVs[kind]; !ok {
		return nil, fmt.Errorf("KV %v not supported", kind)
	}
	switch kind {
	case INMEMORY:
		kv = &InMemory{Db: make(map[string]*Value)}
		break
	}
	return kv, nil
}
