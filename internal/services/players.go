package services

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/antonaby/shortsbattle/game-server/internal/common"
	"github.com/antonaby/shortsbattle/game-server/internal/db"
	"github.com/antonaby/shortsbattle/game-server/internal/db/qg"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

type PlayersService struct {
	txm db.TxManager
	rc  *redis.Client
}

func NewPlayersService(txm db.TxManager, rc *redis.Client) *PlayersService {
	return &PlayersService{
		txm: txm,
		rc:  rc,
	}
}

func (ps *PlayersService) CheckPlayerExistsOrCreate(ctx context.Context, params qg.CreatePlayerParams) (*qg.Player, error) {
	playerKey := fmt.Sprintf("player:%d", params.TgID)

	playerFromRedis, err := ps.rc.HGetAll(ctx, playerKey).Result()
	if err != nil {
		log.Error().Err(err).Msg("can't get player details from Redis")
	} else if len(playerFromRedis) > 0 {
		player, err := mapToPlayer(playerFromRedis)
		if err == nil {
			return player, nil
		}
	}

	player, err := db.WithTxValue(ctx, ps.txm, func(ctx context.Context, tx pgx.Tx) (*qg.Player, error) {
		q := ps.txm.Querier(tx)
		return ps.getPlayerFromDb(ctx, q, params)
	})

	if err != nil {
		return nil, err
	}

	pipe := ps.rc.TxPipeline()
	pipe.HSet(ctx, playerKey, playerToMap(player))
	pipe.Expire(ctx, playerKey, 24*time.Hour)
	_, err = pipe.Exec(ctx)
	if err != nil {
		log.Error().Err(err).Msg("can't save player details to Redis")
	}

	return player, nil
}

func (ps *PlayersService) getPlayerFromDb(ctx context.Context, q qg.Querier, params qg.CreatePlayerParams) (*qg.Player, error) {
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

func mapToPlayer(m map[string]string) (*qg.Player, error) {
	var p qg.Player
	var err error

	if v, ok := m["tg_id"]; ok {
		p.TgID, err = strconv.ParseInt(v, 10, 64)
		if err != nil {
			return nil, err
		}
	}

	if v, ok := m["tg_username"]; ok {
		p.TgUsername = v
	}

	if v, ok := m["tg_language_code"]; ok {
		p.TgLanguageCode = v
	}

	if v, ok := m["created_at"]; ok {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			return nil, err
		}
		p.CreatedAt = pgtype.Timestamptz{Time: t, Valid: true}
	}

	return &p, nil
}

func playerToMap(p *qg.Player) map[string]string {
	m := make(map[string]string)

	m["tg_id"] = strconv.FormatInt(p.TgID, 10)
	m["tg_username"] = p.TgUsername
	m["tg_language_code"] = p.TgLanguageCode

	if p.CreatedAt.Valid {
		m["created_at"] = p.CreatedAt.Time.UTC().Format(time.RFC3339)
	} else {
		m["created_at"] = ""
	}

	return m
}
