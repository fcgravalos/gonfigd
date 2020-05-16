// Package pubsub provides the needed primitives for a Publish-Subscribe system
package pubsub

import (
	"sync"
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
// returns the newly created Subscription object,
// or NoSuchTopicError if the topic is not created yet
func (im *InMemory) Subscribe(topic string) (*Subscription, error) {
	if !im.TopicExists(topic) {
		return nil, NewNoSuchTopicError(topic)
	}
	s := NewSubscription()
	im.Lock()
	im.pubsub[topic][s.ID()] = s.Channel()
	im.Unlock()
	return s, nil
}

// UnSubscribe removes a subscription from a topic
func (im *InMemory) UnSubscribe(topic string, sID string) error {
	if !im.TopicExists(topic) {
		return NewNoSuchTopicError(topic)
	}
	im.Lock()
	close(im.pubsub[topic][sID])
	delete(im.pubsub[topic], sID)
	im.Unlock()
	return nil
}
