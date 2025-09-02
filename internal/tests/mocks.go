package tests

import (
	"context"

	"github.com/antonaby/shortsbattle/game-server/internal/db/qg"
	"github.com/stretchr/testify/mock"
)

type MockPlayerService struct {
	mock.Mock
}

func (mps *MockPlayerService) CheckPlayerExistsOrCreate(ctx context.Context, params qg.CreatePlayerParams) (*qg.Player, error) {
	args := mps.Called(ctx, params)

	var player *qg.Player
	if p := args.Get(0); p != nil {
		player = p.(*qg.Player)
	}

	return player, args.Error(1)
}