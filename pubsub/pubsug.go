// Package pubsub provides the needed primitives for a Publish-Subscribe system
package pubsub

import (
	"github.com/google/uuid"
)

//INMEMORY is an in-memory PubSub implementation
const INMEMORY Kind = "in-memory"

var supportedPubSubs map[string]Kind = map[string]Kind{
	"in-memory": INMEMORY,
}

// Kind is the KV kind
type Kind string

// Subscription type
type Subscription struct {
	id string
	ch chan *Event
}

// ID returns the subscription id
func (s *Subscription) ID() string {
	return s.id
}

// Channel returns the subscription channel
func (s *Subscription) Channel() chan *Event {
	return s.ch
}

// NewSubscription creates a new Subscription
func NewSubscription() *Subscription {
	return &Subscription{
		id: uuid.New().String(),
		ch: make(chan *Event),
	}
}

// PubSub provides an API for a Publish-Subscribe system
type PubSub interface {
	CreateTopic(topic string) error
	DeleteTopic(topic string) error
	TopicExists(topic string) bool
	Publish(topic string, ev *Event) error
	Subscribe(topic string) (*Subscription, error)
	UnSubscribe(topic string, sID string) error
}

// PubSubFromName returns the PubSub Kind from the provided name
// It will return a NotImplementedError otherwise
func PubSubFromName(name string) (Kind, error) {
	kind, ok := supportedPubSubs[name]
	if !ok {
		return kind, NewNotImplementedError(name)
	}
	return kind, nil
}

// NewPubSub returns an implementation of the PubSub interface
func NewPubSub(kind Kind) (PubSub, error) {
	var ps PubSub

	switch kind {
	case INMEMORY:
		ps = &InMemory{pubsub: make(map[string]subscriptions)}
		break
	}
	return ps, nil
}
