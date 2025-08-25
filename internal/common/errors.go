package common

import (
	"fmt"

	"github.com/antonaby/shortsbattle/game-server/internal/db"
)

type ErrorCode int

const (
	ErrorDbUnknown = iota
	ErrorDbNotFound
	ErrorWrongGameState
	ErrorDbConstraintViolation
	ErrorOEmbedFailed
	ErrorRedisStream
	ErrorParse
)

type ServiceError struct {
	Code    ErrorCode
	Message string
	Cause   error
}

func (e ServiceError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("Error %d: %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("Error %d: %s", e.Code, e.Message)
}

func (e ServiceError) Unwrap() error {
	return e.Cause
}

func GetDbErrorCode(err error) ErrorCode {
	if db.IsClass23(err) {
		return ErrorDbConstraintViolation
	}
	if db.IsNoRows(err) {
		return ErrorDbNotFound
	}

	return ErrorDbUnknown
}
