package services

import "time"

type GameConfig struct {
	MaxPlayers      int32
	LobbyState      time.Duration
	SubmittingState time.Duration
	WatchingState   time.Duration
}
