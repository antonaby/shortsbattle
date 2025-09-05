package services

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/antonaby/shortsbattle/game-server/internal/common"
)

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
