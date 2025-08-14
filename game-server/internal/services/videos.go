package services

import (
	"fmt"

	"github.com/antonaby/shortsbattle/game-server/internal/db"
)

type VideosServiceErrorCode int

const (
	VSErrDbError = iota
	VSErrNotFound
)

type VideosServiceError struct {
	Code    VideosServiceErrorCode
	Message string
	Cause   error
}

func (e VideosServiceError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("Error %d: %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("Error %d: %s", e.Code, e.Message)
}

func (e VideosServiceError) Unwrap() error {
	return e.Cause
}

type VideosService struct {
	txm db.TxManager
}

func NewVideosService(txm db.TxManager) *VideosService {
	return &VideosService{
		txm: txm,
	}
}
