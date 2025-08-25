package services

import (
	"fmt"
	"time"
)

type GameConfig struct {
	MaxPlayers        int32
	LobbyState        time.Duration
	LobbyClosedBefore time.Duration
	SubmittingState   time.Duration
	WatchingState     time.Duration
}

func CfChannelName(gameId int64) string {
	return fmt.Sprintf("game_%d", gameId)
}