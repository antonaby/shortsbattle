package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/antonaby/shortsbattle/game-server/internal/db"
	"github.com/jackc/pgx/v5"
)

type PlayersServiceErrorCode int

const (
	PSErrDbError = iota
	PSErrNotFound
)

type PlayersServiceError struct {
	Code    PlayersServiceErrorCode
	Message string
	Cause   error
}

func (e PlayersServiceError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("Error %d: %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("Error %d: %s", e.Code, e.Message)
}

func (e PlayersServiceError) Unwrap() error {
	return e.Cause
}

type PlayersService struct {
	txm db.TxManager
}

func NewPlayersService(txm db.TxManager) *PlayersService {
	return &PlayersService{
		txm: txm,
	}
}

func (g *PlayersService) CreatePlayer(ctx context.Context, params db.CreatePlayerParams) (*db.Player, error) {
	return db.WithTxValue(ctx, g.txm, func(ctx context.Context, tx pgx.Tx) (*db.Player, error) {
		q := g.txm.Querier(tx)
		player, err := q.CreatePlayer(ctx, params)
		if err != nil {
			return nil, PlayersServiceError{
				Code:    PSErrDbError,
				Message: "can't create player",
				Cause:   err,
			}
		}

		return &player, nil
	})
}

func (g *PlayersService) GetPlayer(ctx context.Context, id int64) (*db.Player, error) {
	return db.WithTxValue(ctx, g.txm, func(ctx context.Context, tx pgx.Tx) (*db.Player, error) {
		q := g.txm.Querier(tx)
		player, err := q.GetPlayer(ctx, db.GetPlayerParams{ID: id})
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, PlayersServiceError{
					Code:    PSErrNotFound,
					Message: "player not found",
				}
			}

			return nil, PlayersServiceError{
				Code:    PSErrDbError,
				Message: "can't get player",
				Cause:   err,
			}
		}

		return &player, nil
	})
}
