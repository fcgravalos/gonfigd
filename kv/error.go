package kv

import "fmt"

const (
	KeyNotFound ErrType = "KEY_NOT_FOUND_ERROR"
	Unknown     ErrType = "UNKNOWN_ERROR"
)

type ErrType string

type KeyNotFoundError struct {
	errType ErrType
	key     string
}

func (e KeyNotFoundError) Error() string {
	return fmt.Sprintf("[%s] Key %s not found in KV", e.errType, e.key)
}

func NewKeyNotFoundError(key string) KeyNotFoundError {
	return KeyNotFoundError{errType: KeyNotFound, key: key}
}

func getErrorType(e error) ErrType {
	switch e.(type) {
	case KeyNotFoundError:
		return KeyNotFound
	default:
		return Unknown
	}
}

func isKeyNotFoundError(e error) bool {
	return getErrorType(e) == KeyNotFound
}
