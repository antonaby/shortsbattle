package services

import (
	"context"
	"testing"
	"time"

	"github.com/antonaby/shortsbattle/game-server/internal/db"
	"github.com/antonaby/shortsbattle/game-server/internal/db/qg"
	"github.com/antonaby/shortsbattle/game-server/internal/tests"
	"github.com/ory/dockertest/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type GameManagerTestSuite struct {
	pool        *dockertest.Pool
	pgContainer *dockertest.Resource
	dbManager   *db.DbManager
}

func NewGMTestSuite(t *testing.T) *GameManagerTestSuite {
	pool, err := tests.CreateDockerPool()
	if err != nil {
		t.Fatalf("failed to connect to docker: %v", err)
	}

	pgUser := "admin"
	pgPassword := "123"
	pgDb := "sbtest"

	pgContainer, err := tests.CreatePostgresContainer(t, pool, pgUser, pgPassword, pgDb)
	if err != nil {
		t.Fatalf("failed to create pg container: %v", err)
	}

	dbManager, err := tests.CreateDbManager(pool, pgContainer, pgUser, pgPassword, pgDb)
	if err != nil {
		t.Fatalf("failed to create db manager: %v", err)
	}

	t.Cleanup(func() {
		dbManager.Close()
	})

	return &GameManagerTestSuite{
		pool:        pool,
		pgContainer: pgContainer,
		dbManager:   dbManager,
	}
}

func TestAdvanceGame(t *testing.T) {
	ts := NewGMTestSuite(t)

	theme, err := tests.CreateTestTheme(ts.dbManager, 3)
	if err != nil {
		t.Fatalf("failed to create test theme: %v", err)
	}

	players, err := tests.CreateTestPlayers(ts.dbManager, 2)
	if err != nil {
		t.Fatalf("failed to create test theme: %v", err)
	}

	gm := NewGameManager(ts.dbManager, GameConfig{
		MaxPlayers:                 2,
		MinRemainingBeforeChangeMs: 300,
		MinLobbyState:              1 * time.Second,
		MaxLobbyState:              2 * time.Second,
		LobbyClosedBefore:          5 * time.Second,
		SubmittingState:            60 * time.Second,
		WatchingState:              600 * time.Second,
	})

	t.Run("LobbyWithMaxPlayers", func(t *testing.T) {
		ctx, cancelFunc := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelFunc()

		// 1) Create game
		game, err := db.WithTxVQ(ctx, ts.dbManager, func(ctx context.Context, q qg.Querier) (qg.Game, error) {
			return q.TestCreateGame(ctx, qg.TestCreateGameParams{
				ThemeID:      theme.ID,
				State:        qg.GameStateLobby,
				NextChangeIn: db.ToPgInterval(gm.config.MaxLobbyState),
			})
		})
		if err != nil {
			t.Fatalf("failed to create test game: %v", err)
		}
		upd, err := gm.AdvanceGame(ctx, game.ID)
		if err != nil {
			t.Fatalf("failed to advance test game: %v", err)
		}
		// 2) No time to change state, nPlayer = 0
		require.Nil(t, upd, "update not nil")

		// 3) Add Player 1
		_, err = db.WithTxVQ(ctx, ts.dbManager, func(ctx context.Context, q qg.Querier) (qg.GamePlayer, error) {
			return q.TestAddPlayerToGame(ctx, qg.TestAddPlayerToGameParams{
				GameID:   game.ID,
				PlayerID: players[0].TgID,
			})
		})
		if err != nil {
			t.Fatalf("failed to add player 1 to game: %v", err)
		}

		upd, err = gm.AdvanceGame(ctx, game.ID)
		if err != nil {
			t.Fatalf("failed to advance test game: %v", err)
		}
		// 4) No time to change state, nPlayer = 1
		require.Nil(t, upd, "update not nil")

		// 5) Add Player 2
		_, err = db.WithTxVQ(ctx, ts.dbManager, func(ctx context.Context, q qg.Querier) (qg.GamePlayer, error) {
			return q.TestAddPlayerToGame(ctx, qg.TestAddPlayerToGameParams{
				GameID:   game.ID,
				PlayerID: players[1].TgID,
			})
		})
		if err != nil {
			t.Fatalf("failed to add player 2 to game: %v", err)
		}

		upd, err = gm.AdvanceGame(ctx, game.ID)
		if err != nil {
			t.Fatalf("failed to advance test game: %v", err)
		}
		// 6) No time to change state, nPlayer = 2, but minLobbyTime not exceeded
		require.Nil(t, upd, "update not nil")

		// 7) Fetch game again and check that remaning time reduced
		game, err = db.WithTxVQ(ctx, ts.dbManager, func(ctx context.Context, q qg.Querier) (qg.Game, error) {
			return q.TestGetGameById(ctx, qg.TestGetGameByIdParams{ID: game.ID})
		})
		if err != nil {
			t.Fatalf("failed to fetc game again: %v", err)
		}
		assert.Less(t, game.NextStateChangeAt.Time.Sub(game.StateChangedAt.Time), gm.config.MaxLobbyState)

		// 8) Wait and Advance
		time.Sleep(1 * time.Second)
		upd, err = gm.AdvanceGame(ctx, game.ID)
		if err != nil {
			t.Fatalf("failed to advance test game: %v", err)
		}
		// 9) Game state changed to submitting
		require.NotNil(t, upd)
		assert.Equal(t, qg.GameStateSubmitting, upd.State)
		assert.Equal(t, upd.StateChangeReason, ReasonLobbyFullPlayers)
	})

	t.Run("LobbyTimeout", func(t *testing.T) {
		ctx, cancelFunc := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelFunc()

		// 1) Create game
		game, err := db.WithTxVQ(ctx, ts.dbManager, func(ctx context.Context, q qg.Querier) (qg.Game, error) {
			return q.TestCreateGame(ctx, qg.TestCreateGameParams{
				ThemeID:      theme.ID,
				State:        qg.GameStateLobby,
				NextChangeIn: db.ToPgInterval(gm.config.MaxLobbyState),
			})
		})
		if err != nil {
			t.Fatalf("failed to create test game: %v", err)
		}
		upd, err := gm.AdvanceGame(ctx, game.ID)
		if err != nil {
			t.Fatalf("failed to advance test game: %v", err)
		}
		// 2) No time to change state, nPlayer = 0, no timeout
		require.Nil(t, upd, "update not nil")

		// 3) Wait and Advance
		time.Sleep(2 * time.Second)
		upd, err = gm.AdvanceGame(ctx, game.ID)
		if err != nil {
			t.Fatalf("failed to advance test game: %v", err)
		}
		// 4) Game state changed to submitting
		require.NotNil(t, upd)
		assert.Equal(t, qg.GameStateSubmitting, upd.State)
		assert.Equal(t, upd.StateChangeReason, ReasonLobbyTimeout)
	})
}
