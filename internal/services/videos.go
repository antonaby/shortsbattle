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
	"github.com/antonaby/shortsbattle/game-server/internal/models"
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

func (vs *VideosService) CreateVideo(ctx context.Context, params models.SubmitVideoRequest) (*qg.Video, error) {
	return db.WithTxValue(ctx, vs.txm, func(ctx context.Context, tx pgx.Tx) (*qg.Video, error) {
		oembed, err := fetchOEmbed(ctx, params.VideoUrl, "") // TODO: add Instagram Access Token
		if err != nil {
			return nil, common.ServiceError{
				Code:    common.ErrorOEmbedFailed,
				Message: "failed to submit video",
				Cause:   err,
			}
		}

		q := vs.txm.Querier(tx)
		video, err := q.CreateVideo(ctx, qg.CreateVideoParams{
			PlayerID: params.PlayerID,
			VideoUrl: params.VideoUrl,
			Oembed:   oembed,
		})
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

func (vs *VideosService) GetVideo(ctx context.Context, videoId int64) (*qg.Video, error) {
	return db.WithTxValue(ctx, vs.txm, func(ctx context.Context, tx pgx.Tx) (*qg.Video, error) {
		q := vs.txm.Querier(tx)
		video, err := q.GetVideo(ctx, qg.GetVideoParams{ID: videoId})
		if err != nil {
			return nil, common.ServiceError{
				Code:    common.GetDbErrorCode(err),
				Message: "failed to fetch video",
				Cause:   err,
			}
		}

		return &video, nil
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
