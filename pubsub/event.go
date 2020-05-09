package pubsub

import (
	"fmt"
	"time"
)

const (
	// ConfigCreated EventType indicates that a new config has been created
	ConfigCreated EventType = iota
	// ConfigUpdated EventType indicates that the config has changed
	ConfigUpdated
	// ConfigDeleted EventType indicates that the config has been deleted
	ConfigDeleted
)

// EventType is type of configuration change Event
type EventType uint8

// String returns the string version of EventType
func (t EventType) String() string {
	switch {
	case t == ConfigCreated:
		return "CONFIG_CREATED"
	case t == ConfigUpdated:
		return "CONFIG_UPDATED"
	case t == ConfigDeleted:
		return "CONFIG_DELETED"
	}
	return "UNKNOWN"
}

// Event data structure
type Event struct {
	// Event kind
	kind EventType
	// The configuration updated
	configPath string
	// Creation timestamp
	createdAt time.Time
}

// CreateEvent returns a a pointer to Event type
// kind EventType, the kind of Event
// configPath string, the config path
func CreateEvent(kind EventType, configPath string) *Event {
	return &Event{kind: kind, configPath: configPath, createdAt: time.Now()}
}

// Kind returns the EventType
func (ev *Event) Kind() EventType {
	return ev.kind
}

// ConfigPath returns the path to
func (ev *Event) ConfigPath() string {
	return ev.configPath
}

// CreatedAt returns the timestamp when the Event was created
func (ev *Event) CreatedAt() time.Time {
	return ev.createdAt
}

// String returns a new string with the Event information
func (ev *Event) String() string {
	return fmt.Sprintf("[%s] - %s: %s", ev.createdAt.String(), ev.kind.String(), ev.configPath)
}
