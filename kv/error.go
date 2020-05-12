package kv

import "fmt"

const (
	KeyNotFound    ErrType = "KEY_NOT_FOUND_ERROR"
	Compression    ErrType = "COMPRESSION_ERROR"
	NotImplemented ErrType = "NOT_IMPLEMENTED_ERROR"
	Unknown        ErrType = "UNKNOWN_ERROR"
)

type ErrType string

type KeyNotFoundError struct {
	errType ErrType
	key     string
}

type CompressionError struct {
	errType ErrType
	data    string
	err     error
}

type NotImplementedError struct {
	errType ErrType
	kvImpl  string
}

func getErrorType(e error) ErrType {
	switch e.(type) {
	case KeyNotFoundError:
		return KeyNotFound
	case CompressionError:
		return Compression
	case NotImplementedError:
		return NotImplemented
	default:
		return Unknown
	}
}

func IsKeyNotFoundError(e error) bool {
	return getErrorType(e) == KeyNotFound
}

func IsCompressionError(e error) bool {
	return getErrorType(e) == Compression
}

func IsNotImplementedError(e error) bool {
	return getErrorType(e) == NotImplemented
}

func (e KeyNotFoundError) Error() string {
	return fmt.Sprintf("[%s] Key %s not found in KV", e.errType, e.key)
}

func (e CompressionError) Error() string {
	return fmt.Sprintf("[%s] gzip operation failed for data %s", e.errType, e.data)
}

func (e NotImplementedError) Error() string {
	return fmt.Sprintf("[%s] %s is not a supported implementation of KV interface", e.errType, e.kvImpl)
}

func NewKeyNotFoundError(key string) KeyNotFoundError {
	return KeyNotFoundError{errType: KeyNotFound, key: key}
}

func NewCompressionError(data []byte, err error) CompressionError {
	return CompressionError{errType: Compression, data: string(data), err: err}
}

func NewNotImplementedError(impl string) NotImplementedError {
	return NotImplementedError{errType: NotImplemented, kvImpl: impl}
}
