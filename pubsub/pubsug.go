// Package pubsub provides the needed primitives for a Publish-Subscribe system
package pubsub

import (
	"fmt"
)

//INMEMORY is an in-memory PubSub implementation
const INMEMORY Kind = "in-memory"

var supportedPubSubs map[Kind]struct{}

func init() {
	supportedPubSubs = map[Kind]struct{}{INMEMORY: struct{}{}}
}

// Kind is the KV kind
type Kind string

// PubSub provides an API for a Publish-Subscribe system
type PubSub interface {
	CreateTopic(topic string) error
	DeleteTopic(topic string) error
	TopicExists(topic string) bool
	Publish(topic string, ev *Event) error
	Subscribe(topic string) (string, chan *Event)
	UnSubscribe(topic string, sID string) error
}

// NewPubSub returns an implementation of the PubSub interface
func NewPubSub(kind Kind) (PubSub, error) {
	var ps PubSub
	if _, ok := supportedPubSubs[kind]; !ok {
		return nil, fmt.Errorf("PubSub %v not supported", kind)
	}
	switch kind {
	case INMEMORY:
		ps = &InMemory{pubsub: make(map[string]subscriptions)}
		break
	}
	return ps, nil
}
