// Package pubsub provides the needed primitives for a Publish-Subscribe system
package pubsub

import (
	"sync"

	"github.com/google/uuid"
)

type subscriptions map[string]chan *Event

// InMemory is the data structure implementing the PubSub interface
type InMemory struct {
	sync.Mutex
	pubsub map[string]subscriptions
}

// CreateTopic creates a new topic from string
func (im *InMemory) CreateTopic(topic string) error {
	im.Lock()
	im.pubsub[topic] = make(subscriptions, 0)
	im.Unlock()
	return nil
}

// DeleteTopic deletes a given topic
func (im *InMemory) DeleteTopic(topic string) error {
	im.Lock()
	_, ok := im.pubsub[topic]
	if ok {
		delete(im.pubsub, topic)
	}
	im.Unlock()
	return nil
}

// TopicExists checks whether or not a topic exists
func (im *InMemory) TopicExists(topic string) bool {
	_, ok := im.pubsub[topic]
	return ok
}

// Publish injects a new *Event into a topic
func (im *InMemory) Publish(topic string, ev *Event) error {
	for _, sCh := range im.pubsub[topic] {
		sCh <- ev
	}
	return nil
}

// Subscribe adds a new subscription to a topic,
// returns the subscription ID and a subscription channel
func (im *InMemory) Subscribe(topic string) (string, chan *Event) {
	sID := uuid.New().String()
	sCh := make(chan *Event)
	im.Lock()
	im.pubsub[topic][sID] = sCh
	im.Unlock()
	return sID, sCh
}

// UnSubscribe removes a subscription from a topic
func (im *InMemory) UnSubscribe(topic string, sID string) error {
	im.Lock()
	close(im.pubsub[topic][sID])
	delete(im.pubsub[topic], sID)
	im.Unlock()
	return nil
}
