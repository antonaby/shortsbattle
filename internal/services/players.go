package services

import (
	"context"

	"github.com/antonaby/shortsbattle/game-server/internal/common"
	"github.com/antonaby/shortsbattle/game-server/internal/db"
	"github.com/antonaby/shortsbattle/game-server/internal/db/qg"
	"github.com/jackc/pgx/v5"
	"github.com/redis/go-redis/v9"
)

type PlayersService struct {
	txm db.TxManager
	rc *redis.Client
}

func NewPlayersService(txm db.TxManager, rc *redis.Client) *PlayersService {
	return &PlayersService{
		txm: txm,
		rc: rc,
	}
}

func (ps *PlayersService) CheckPlayerExistsOrCreate(ctx context.Context, params qg.CreatePlayerParams) (*qg.Player, error) {
	return db.WithTxValue(ctx, ps.txm, func(ctx context.Context, tx pgx.Tx) (*qg.Player, error) {
		q := ps.txm.Querier(tx)
		player, err := q.GetPlayerByTgId(ctx, qg.GetPlayerByTgIdParams{TgID: params.TgID})
		if err != nil {
			if db.IsNoRows(err) {
				return ps.createPlayer(ctx, q, params)
			}

			return nil, common.ServiceError{
				Code:    common.ErrorDb,
				Message: "failed to get player",
				Cause:   err,
			}
		}

		return &player, nil
	})
}

func (ps *PlayersService) createPlayer(ctx context.Context, q qg.Querier, params qg.CreatePlayerParams) (*qg.Player, error) {
	player, err := q.CreatePlayer(ctx, params)
	if err != nil {
		return nil, common.ServiceError{
				Code:    common.ErrorDb,
				Message: "failed to create player",
				Cause:   err,
			} 
	}

	return &player, nil
}
