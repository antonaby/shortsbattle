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
)

func vsError(code common.ErrorCode, msg string, err error) error {
	return common.ServiceError{
		Code:    code,
		Message: fmt.Sprintf("video service: %s", msg),
		Cause:   err,
	}
}

func vsDbError(msg string, err error) error {
	return vsError(
		common.GetDbErrorCode(err),
		msg,
		err,
	)
}

type VideoService struct {
	txm db.TxManager
}

func NewVideosService(txm db.TxManager) *VideoService {
	return &VideoService{
		txm: txm,
	}
}

func (vs *VideoService) AddVideo(ctx context.Context, playerId int64, videoUrl string) (*qg.Video, error) {
	oembed, err := fetchOEmbed(ctx, videoUrl)
	if err != nil {
		return nil, err
	}

	return db.WithTxVQ(ctx, vs.txm, func(ctx context.Context, q qg.Querier) (*qg.Video, error) {
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
		return nil, vsDbError("failed to create video", err)
	}

	return &video, nil
}

// TODO: add pagination and search
func (vs *VideoService) GetVideosByPlayer(ctx context.Context, playerId int64, query string) ([]qg.Video, error) {
	return db.WithTxVQ(ctx, vs.txm, func(ctx context.Context, q qg.Querier) ([]qg.Video, error) {
		videos, err := q.GetVideosByPlayer(ctx, playerId)
		if err != nil {
			return nil, vsDbError("failed to fetch video", err)
		}

		if len(videos) == 0 {
			videos = []qg.Video{}
		}

		return videos, nil
	})
}

func oembedEndpoint(videoURL string) (string, error) {
	u, err := url.Parse(videoURL)
	if err != nil {
		return "", vsError(common.ErrorOEmbedFailed, "failed to parse video url", err)
	}

	host := strings.ToLower(u.Host)
	escapedUrl := url.QueryEscape(videoURL)

	switch {
	case strings.Contains(host, "youtube.com"), strings.Contains(host, "youtu.be"):
		return "https://www.youtube.com/oembed?format=json&url=" + escapedUrl, nil
	case strings.Contains(host, "tiktok.com"):
		return "https://www.tiktok.com/oembed?url=" + escapedUrl, nil
	}

	return "", vsError(common.ErrorOEmbedFailed, fmt.Sprintf("unsupported host %s", host), err)
}

func fetchOEmbed(ctx context.Context, videoURL string) ([]byte, error) {
	endpoint, err := oembedEndpoint(videoURL)
	if err != nil {
		return nil, err
	}

	req, _ := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, vsError(common.ErrorOEmbedFailed, "failed to get oembed data", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return nil, vsError(common.ErrorOEmbedFailed, fmt.Sprintf("oembed status %d: %s", resp.StatusCode, string(b)), err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, vsError(common.ErrorOEmbedFailed, "failed to read oembed data", err)
	}

	return body, nil
}
