package services

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/antonaby/shortsbattle/game-server/internal/common"
	"github.com/antonaby/shortsbattle/game-server/internal/db"
	"github.com/antonaby/shortsbattle/game-server/internal/db/qg"
	"github.com/jackc/pgx/v5"
)

type VideoService struct {
	txm db.TxManager
}

func NewVideosService(txm db.TxManager) *VideoService {
	return &VideoService{
		txm: txm,
	}
}

func (vs *VideoService) AddVideo(ctx context.Context, playerId int64, videoUrl string) (*qg.Video, error) {
	oembed, err := fetchOEmbed(ctx, videoUrl, "")
	if err != nil {
		return nil, common.ServiceError{
			Code:    common.ErrorOEmbedFailed,
			Message: "failed to get oembed data",
			Cause:   err,
		}
	}

	return db.WithTxValue(ctx, vs.txm, func(ctx context.Context, tx pgx.Tx) (*qg.Video, error) {
		q := vs.txm.Querier(tx)
		return vs.createVideo(ctx, q, playerId, videoUrl, oembed)
	})
}

func (vs *VideoService) createVideo(ctx context.Context, q qg.Querier, playerId int64, videoUrl string, oembed []byte) (*qg.Video, error) {
	video, err := q.AddVideoToPlayer(ctx, qg.AddVideoToPlayerParams{
		PlayerID: playerId,
		VideoUrl: videoUrl,
		Oembed:   oembed,
	})

	if err != nil {
		return nil, common.ServiceError{
			Code:    common.GetDbErrorCode(err),
			Message: "failed to create video",
			Cause:   err,
		}
	}

	return &video, nil
}

// TODO: add pagination and search
func (vs *VideoService) GetVideosByPlayer(ctx context.Context, playerId int64) ([]qg.Video, error) {
	return db.WithTxValue(ctx, vs.txm, func(ctx context.Context, tx pgx.Tx) ([]qg.Video, error) {
		q := vs.txm.Querier(tx)
		videos, err := q.GetVideosByPlayer(ctx, playerId)
		if err != nil {
			return nil, common.ServiceError{
				Code:    common.GetDbErrorCode(err),
				Message: "failed to fetch video",
				Cause:   err,
			}
		}

		if len(videos) == 0 {
			videos = []qg.Video{}
		}

		return videos, nil
	})
}

func oembedEndpoint(videoURL, igToken string) (string, error) {
	u, err := url.Parse(videoURL)
	if err != nil {
		return "", common.ServiceError{
			Code:    common.ErrorOEmbedFailed,
			Message: "failed to fetch oembed data",
			Cause:   err,
		}
	}

	host := strings.ToLower(u.Host)
	escapedUrl := url.QueryEscape(videoURL)

	switch {
	case strings.Contains(host, "youtube.com"), strings.Contains(host, "youtu.be"):
		return "https://www.youtube.com/oembed?format=json&url=" + escapedUrl, nil
	case strings.Contains(host, "tiktok.com"):
		return "https://www.tiktok.com/oembed?url=" + escapedUrl, nil
	case strings.Contains(host, "instagram.com"):
		if igToken == "" {
			return "", common.ServiceError{
				Code:    common.ErrorOEmbedFailed,
				Message: "instagram access token not provided",
				Cause:   err,
			}
		}
		return "https://graph.facebook.com/v21.0/instagram_oembed?url=" + escapedUrl + "&access_token=" + url.QueryEscape(igToken), nil
	}

	return "", common.ServiceError{
		Code:    common.ErrorOEmbedFailed,
		Message: fmt.Sprintf("unsupported host %s", host),
		Cause:   err,
	}
}

func fetchOEmbed(ctx context.Context, videoURL, igToken string) ([]byte, error) {
	endpoint, err := oembedEndpoint(videoURL, igToken)
	if err != nil {
		return nil, err
	}

	req, _ := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, common.ServiceError{
			Code:    common.ErrorOEmbedFailed,
			Message: "failed to fetch oembed data",
			Cause:   err,
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return nil, common.ServiceError{
			Code:    common.ErrorOEmbedFailed,
			Message: fmt.Sprintf("oembed status %d: %s", resp.StatusCode, string(b)),
			Cause:   err,
		}
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, common.ServiceError{
			Code:    common.ErrorOEmbedFailed,
			Message: "failed to fetch oembed data",
			Cause:   err,
		}
	}

	return body, nil
}
