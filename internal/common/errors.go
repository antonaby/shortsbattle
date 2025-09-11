package common

import (
	"fmt"

	"github.com/antonaby/shortsbattle/game-server/internal/db"
)

type ErrorCode int

const (
	ErrorUnknown = iota
	ErrorNotFound
	ErrorBadData
	ErrorWrongGameStage
	ErrorConstraintViolation
	ErrorOEmbedFailed
	ErrorRedisStream
	ErrorRedis
	ErrorParse
	ErrorMarshal
	ErrorEnqueue
	ErrorTgInitData
	ErrorJWT
	ErrorJWK
	ErrorForbidden
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
		return ErrorConstraintViolation
	}
	if db.IsNoRows(err) {
		return ErrorNotFound
	}
	if db.IsClass22(err) {
		return ErrorBadData
	}

	return ErrorUnknown
}
