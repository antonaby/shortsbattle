package services

import (
	"context"

	"github.com/antonaby/shortsbattle/game-server/internal/common"
	"github.com/antonaby/shortsbattle/game-server/internal/db"
	"github.com/antonaby/shortsbattle/game-server/internal/db/qg"
	"github.com/jackc/pgx/v5"
)

type VideosService struct {
	txm db.TxManager
}

func NewVideosService(txm db.TxManager) *VideosService {
	return &VideosService{
		txm: txm,
	}
}

func (vs *VideosService) CreateVideo(ctx context.Context, params qg.CreateVideoParams) (*qg.Video, error) {
	return db.WithTxValue(ctx, vs.txm, func(ctx context.Context, tx pgx.Tx) (*qg.Video, error) {
		q := vs.txm.Querier(tx)
		video, err := q.CreateVideo(ctx, params)
		if err != nil {
			return nil, common.ServiceError{
				Code:    common.GetDbErrorCode(err),
				Message: "failed to submit video",
				Cause:   err,
			}
		}

		return &video, nil
	})
}
