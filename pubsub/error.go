package pubsub

import "fmt"

const (
	NoSuchTopic    ErrType = "NO_SUCH_TOPIC_ERROR"
	NotImplemented ErrType = "NOT_IMPLEMENTED_ERROR"
	Unknown        ErrType = "UNKNOWN_ERROR"
)

type ErrType string

type NoSuchTopicError struct {
	errType ErrType
	topic   string
}

type NotImplementedError struct {
	errType ErrType
	psImpl  string
}

func getErrorType(e error) ErrType {
	switch e.(type) {
	case NoSuchTopicError:
		return NoSuchTopic
	case NotImplementedError:
		return NotImplemented
	default:
		return Unknown
	}
}

func IsNoSuchTopicError(e error) bool {
	return getErrorType(e) == NoSuchTopic
}

func IsNotImplementedError(e error) bool {
	return getErrorType(e) == NotImplemented
}

func (e NoSuchTopicError) Error() string {
	return fmt.Sprintf("[%s] Topic %s does not exist", e.errType, e.topic)
}

func (e NotImplementedError) Error() string {
	return fmt.Sprintf("[%s] %s is not a supported implementation of PubSub interface", e.errType, e.psImpl)
}

func NewNoSuchTopicError(topic string) NoSuchTopicError {
	return NoSuchTopicError{errType: NoSuchTopic, topic: topic}
}

func NewNotImplementedError(impl string) NotImplementedError {
	return NotImplementedError{errType: NotImplemented, psImpl: impl}
}
