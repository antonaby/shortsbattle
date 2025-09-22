package models

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/antonaby/shortsbattle/game-server/internal/common"
	"github.com/antonaby/shortsbattle/game-server/internal/db/qg"
	"github.com/jackc/pgx/v5/pgtype"
)

type ThemeRound struct {
	RoundN      int32  `json:"round_n"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

type ThemeWithRounds struct {
	ID          int64        `json:"id"`
	Title       string       `json:"title"`
	Description string       `json:"description"`
	Mode        qg.GameMode  `json:"mode"`
	Rounds      []ThemeRound `json:"rounds,omitempty"`
}

type LikeDislikeVoteValue string

const (
	LikeValue    LikeDislikeVoteValue = "like"
	DislikeValue LikeDislikeVoteValue = "dislike"
)

type LikeDislikeVote struct {
	Value LikeDislikeVoteValue `json:"value" validate:"required,oneof=like dislike"`
}

type LikeDislikeVideoResult struct {
	GameVideoID int64 `json:"game_video_id"`
	RoundN      int32 `json:"round_n"`
	AuthorID    int64 `json:"author_id"`
	Likes       int   `json:"likes"`
	Dislikes    int   `json:"dislikes"`
}

type LikeDislikeFinalResult struct {
	Results []LikeDislikeVideoResult `json:"results"`
}

type SubmitGameVideo struct {
	ID          int64              `json:"id"`
	GameID      int64              `json:"game_id"`
	PlayerID    int64              `json:"player_id"`
	RoundN      int32              `json:"round_n"`
	SubmittedAt pgtype.Timestamptz `json:"submitted_at"`
	Video       qg.Video           `json:"video"`
}

type MessageType string

const (
	GameDetailsMsg MessageType = "details"
	GameUpdateMsg  MessageType = "stage_updated"
)

type StageChangeReason string

const (
	ReasonLobbyFull                 StageChangeReason = "lobby-full"
	ReasonLobbyTimeout              StageChangeReason = "lobby-timeout"
	ReasonLobbyFullTimeout          StageChangeReason = "lobby-full-timeout"
	ReasonSubmitAll                 StageChangeReason = "submit-all"
	ReasonSubmitTimeout             StageChangeReason = "submit-timeout"
	ReasinSubmitCompleteTimeout     StageChangeReason = "submit-complete-timeout"
	ReasonWatchAll                  StageChangeReason = "watch-all"
	ReasonWatchTimeout              StageChangeReason = "watch-timeout"
	ReasonWatchCompleteNextRound    StageChangeReason = "watch-complete-next-round"
	ReasonWatchCompleteGameComplete StageChangeReason = "watch-complete-game-complete"
)

type GameUpdate struct {
	GameID            int64              `json:"id"`
	PlayerMode        qg.PlayerGameMode  `json:"player_mode,omitempty"`
	GameMode          qg.GameMode        `json:"game_mode,omitempty"`
	MsgType           MessageType        `json:"msg_type"`
	Stage             qg.GameStage       `json:"stage"`
	StateChangeReason *StageChangeReason `json:"state_change_reason,omitempty"`
	RoundN            int32              `json:"round"`
	StateChangedAt    pgtype.Timestamptz `json:"state_changed_at"`
	RemainingMs       int64              `json:"remaining_ms"`
	Theme             *ThemeWithRounds   `json:"theme,omitempty"`
	Result            *json.RawMessage   `json:"result,omitempty"`
}

func GetCfChannelName(gameId int64) string {
	return fmt.Sprintf("game_%d", gameId)
}

func ParseCfChannelName(channel string) (int64, error) {
	const prefix = "game_"
	if !strings.HasPrefix(channel, prefix) {
		return 0, common.ServiceError{
			Code:    common.ErrorParse,
			Message: "failed to parse channel name",
		}
	}

	idStr := strings.TrimPrefix(channel, prefix)
	gameId, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return 0, common.ServiceError{
			Code:    common.ErrorParse,
			Message: "failed to parse game id",
			Cause:   err,
		}
	}

	return gameId, nil
}
