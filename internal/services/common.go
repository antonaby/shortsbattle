package services

import "time"

type GameConfig struct {
	MaxPlayers int32
	LobbyState time.Duration
}