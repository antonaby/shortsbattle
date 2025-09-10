package services

import (
	"context"
	"testing"
	"time"

	"github.com/antonaby/shortsbattle/game-server/internal/common"
	"github.com/antonaby/shortsbattle/game-server/internal/db/qg"
	"github.com/antonaby/shortsbattle/game-server/internal/tests"
	"github.com/stretchr/testify/require"
)

func TestGameActions(t *testing.T) {
	ts := tests.NewDockerTestSuite(t)

	gm := NewGameManager(ts.DBManager, GameConfig{
		MaxPlayers:                 2,
		MinRemainingBeforeChangeMs: 300,
		MinLobbyState:              1 * time.Second,
		MaxLobbyState:              4 * time.Second,
		LobbyClosedBefore:          2 * time.Second,
		SubmittingState:            60 * time.Second,
		WatchingState:              600 * time.Second,
	})

	vs := NewVideosService(ts.DBManager)

	t.Run("JoinGame", func(t *testing.T) {
		ctx, cancelFunc := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelFunc()

		theme, err := tests.CreateTestTheme(ctx, ts.DBManager, 3)
		if err != nil {
			t.Fatalf("failed to create test theme: %v", err)
		}

		players, err := tests.CreateTestPlayers(ctx, ts.DBManager, 3, 0)
		if err != nil {
			t.Fatalf("failed to create test theme: %v", err)
		}

		// player 1 joins game
		gameIdPlayer1Attempt1, err := gm.JoinGame(ctx, theme.ID, players[0].TgID, qg.PlayerGameModeSubmitAndVote)
		if err != nil {
			t.Fatalf("failed to join game (player 1): %v", err)
		}
		// player 1 should join same game
		gameIdPlayer1Attempt2, err := gm.JoinGame(ctx, theme.ID, players[0].TgID, qg.PlayerGameModeSubmitAndVote)
		if err != nil {
			t.Fatalf("failed to join game (player 1): %v", err)
		}

		// check that the game is the same
		require.Equal(t, gameIdPlayer1Attempt1, gameIdPlayer1Attempt2)

		// player 2 should join same game
		gameIdPlayer2, err := gm.JoinGame(ctx, theme.ID, players[1].TgID, qg.PlayerGameModeSubmitAndVote)
		if err != nil {
			t.Fatalf("failed to join game (player 2): %v", err)
		}

		// check that the game is the same
		require.Equal(t, gameIdPlayer1Attempt2, gameIdPlayer2)

		// wait until lobby is closed (LobbyClosedBefore)
		time.Sleep(2 * time.Second)
		gameIdPlayer1Attempt3, err := gm.JoinGame(ctx, theme.ID, players[0].TgID, qg.PlayerGameModeSubmitAndVote)
		if err != nil {
			t.Fatalf("failed to join game (player 1): %v", err)
		}

		// should be the same game as player already in
		require.Equal(t, gameIdPlayer1Attempt2, gameIdPlayer1Attempt3)

		// player 3 should join new game
		gameIdPlayer3, err := gm.JoinGame(ctx, theme.ID, players[2].TgID, qg.PlayerGameModeSubmitAndVote)
		if err != nil {
			t.Fatalf("failed to join game (player 2): %v", err)
		}

		// should be new game as lobby closed (LobbyClosedBefore)
		require.NotEqual(t, gameIdPlayer3, gameIdPlayer1Attempt3)

		// get game status for the previous game
		gameSatatus, err := tests.GetGameStatus(ctx, ts.DBManager, gameIdPlayer1Attempt3)
		if err != nil {
			t.Fatalf("failed to get game status (player 1): %v", err)
		}

		// should still be in the lobby state
		require.Equal(t, qg.GameStageLobby, gameSatatus.Stage)
		require.Greater(t, gameSatatus.RemainingMs, int64(0))
	})

	t.Run("SubmitVideo", func(t *testing.T) {
		ctx, cancelFunc := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelFunc()

		theme, err := tests.CreateTestTheme(ctx, ts.DBManager, 3)
		if err != nil {
			t.Fatalf("failed to create test theme: %v", err)
		}

		players, err := tests.CreateTestPlayers(ctx, ts.DBManager, 3, 10)
		if err != nil {
			t.Fatalf("failed to create test theme: %v", err)
		}

		// add 3 video to player 1
		video1, err := vs.AddVideo(ctx, players[0].TgID, "https://www.youtube.com/shorts/GkTZHFyHi1c")
		if err != nil {
			t.Fatalf("failed to add video (player 1): %v", err)
		}
		_, err = vs.AddVideo(ctx, players[0].TgID, "https://www.youtube.com/shorts/GkTZHFyHi1c")
		if err != nil {
			t.Fatalf("failed to add video (player 1): %v", err)
		}
		video3, err := vs.AddVideo(ctx, players[0].TgID, "https://www.youtube.com/shorts/zy7xd4zOl7s")
		if err != nil {
			t.Fatalf("failed to add video (player 1): %v", err)
		}

		videos, err := vs.GetVideosByPlayer(ctx, players[0].TgID)
		if err != nil {
			t.Fatalf("failed to get videos (player 1): %v", err)
		}
		// check that 2 video has been added (as 1 and 2 video are the same)
		require.Equal(t, 2, len(videos))
		require.Equal(t, video1.ID, videos[0].ID)
		require.Equal(t, video3.ID, videos[1].ID)

		// create game and add player to it
		game1, err := tests.CreateGameWithStage(ctx, ts.DBManager, theme.ID, qg.GameStageSubmit)
		if err != nil {
			t.Fatalf("failed to create game: %v", err)
		}
		_, err = tests.AddPlayerToGame(ctx, ts.DBManager, game1.ID, players[0].TgID, qg.PlayerGameModeSubmitAndVote)
		if err != nil {
			t.Fatalf("failed to add player to game (player 1): %v", err)
		}
		_, err = tests.AddPlayerToGame(ctx, ts.DBManager, game1.ID, players[1].TgID, qg.PlayerGameModeSubmitAndVote)
		if err != nil {
			t.Fatalf("failed to add player to game (player 2): %v", err)
		}

		// add existing video 1
		_, gameVideo1, err := gm.SubmitExistingVideo(ctx, game1.ID, video1.ID, players[0].TgID, 1)
		if err != nil {
			t.Fatalf("failed to submit video (player 1): %v", err)
		}
		// check that video has been added to the same game and player
		require.Equal(t, game1.ID, gameVideo1.GameID)
		require.Equal(t, players[0].TgID, gameVideo1.PlayerID)

		// add existing video 3
		_, gameVideo2, err := gm.SubmitExistingVideo(ctx, game1.ID, video3.ID, players[0].TgID, 1)
		if err != nil {
			t.Fatalf("failed to submit video (player 1): %v", err)
		}

		// check that video has been updated not added
		require.Equal(t, gameVideo1.ID, gameVideo2.ID)
		require.NotEqual(t, gameVideo1.VideoID, gameVideo2.VideoID)

		// add new video
		_, gameVideo3, err := gm.SubmitNewVideo(ctx, game1.ID, "https://www.youtube.com/shorts/wHZyy_G0aYU", players[0].TgID, 1)
		if err != nil {
			t.Fatalf("failed to submit video (player 1): %v", err)
		}
		// check that video has been updated not added
		require.Equal(t, gameVideo1.ID, gameVideo3.ID)
		require.NotEqual(t, gameVideo1.VideoID, gameVideo3.VideoID)

		// add new video, but the one that already exist in DB
		_, gameVideo4, err := gm.SubmitNewVideo(ctx, game1.ID, "https://www.youtube.com/shorts/GkTZHFyHi1c", players[0].TgID, 1)
		if err != nil {
			t.Fatalf("failed to submit video (player 1): %v", err)
		}
		// check that video has been updated not added, the video is the same as video 1
		require.Equal(t, gameVideo1.ID, gameVideo4.ID)
		require.Equal(t, video1.ID, gameVideo4.VideoID)

		// add video for player 2
		_, gameVideo5, err := gm.SubmitNewVideo(ctx, game1.ID, "https://www.youtube.com/shorts/GkTZHFyHi1c", players[1].TgID, 1)
		if err != nil {
			t.Fatalf("failed to submit video (player 2): %v", err)
		}

		// should be different game vidoes but the same video
		require.NotEqual(t, gameVideo4.ID, gameVideo5.ID)
		require.NotEqual(t, gameVideo4.PlayerID, gameVideo5.PlayerID)
		require.Equal(t, gameVideo4.VideoID, gameVideo5.VideoID)

		// player 2 can't submit video owned by player 1
		_, _, notFoundError := gm.SubmitExistingVideo(ctx, game1.ID, video3.ID, players[1].TgID, 1)
		require.NotNil(t, notFoundError)
		notFoundServiceErr := notFoundError.(common.ServiceError)
		require.Equal(t, common.ErrorDbNotFound, int(notFoundServiceErr.Code))

		// player 3 not in the game, so it can't submit video
		_, _, forbiddenErr := gm.SubmitNewVideo(ctx, game1.ID, "https://www.youtube.com/shorts/GkTZHFyHi1c", players[2].TgID, 1)
		require.NotNil(t, forbiddenErr)
		forbiddenServiceErr := forbiddenErr.(common.ServiceError)
		require.Equal(t, common.ErrorForbidden, int(forbiddenServiceErr.Code))

		// chnage game stage to watch to be able to get videos to watch
		_, err = tests.ChangeGameStage(ctx, ts.DBManager, game1.ID, qg.GameStageWatch)
		if err != nil {
			t.Fatalf("failed to chnage game stage: %v", err)
		}

		// get videos for player 1
		videosToWatchP1, err := gm.GetVideosToWatch(ctx, game1.ID, players[0].TgID, 1)
		if err != nil {
			t.Fatalf("failed to get videos to watch (player 1): %v", err)
		}
		// should be videos from player 2
		require.Equal(t, 1, len(videosToWatchP1))
		require.Equal(t, gameVideo5.VideoID, videosToWatchP1[0].VideoID)

		// get videos for player 2
		videosToWatchP2, err := gm.GetVideosToWatch(ctx, game1.ID, players[1].TgID, 1)
		if err != nil {
			t.Fatalf("failed to get videos to watch (player 1): %v", err)
		}
		// should be videos from player 1
		require.Equal(t, 1, len(videosToWatchP2))
		require.Equal(t, gameVideo4.VideoID, videosToWatchP2[0].VideoID)
	})
}

func TestAdvanceGame(t *testing.T) {
	// ts := NewGMTestSuite(t)

	// theme, err := tests.CreateTestTheme(ts.dbManager, 3)
	// if err != nil {
	// 	t.Fatalf("failed to create test theme: %v", err)
	// }

	// players, err := tests.CreateTestPlayers(ts.dbManager, 2)
	// if err != nil {
	// 	t.Fatalf("failed to create test theme: %v", err)
	// }

	// gm := NewGameManager(ts.dbManager, GameConfig{
	// 	MaxPlayers:                 2,
	// 	MinRemainingBeforeChangeMs: 300,
	// 	MinLobbyState:              1 * time.Second,
	// 	MaxLobbyState:              2 * time.Second,
	// 	LobbyClosedBefore:          5 * time.Second,
	// 	SubmittingState:            60 * time.Second,
	// 	WatchingState:              600 * time.Second,
	// })

	// t.Run("LobbyWithMaxPlayers", func(t *testing.T) {
	// 	ctx, cancelFunc := context.WithTimeout(context.Background(), 5*time.Second)
	// 	defer cancelFunc()

	// 	// 1) Create game
	// 	game, err := db.WithTxVQ(ctx, ts.dbManager, func(ctx context.Context, q qg.Querier) (qg.Game, error) {
	// 		return q.TestCreateGame(ctx, qg.TestCreateGameParams{
	// 			ThemeID:      theme.ID,
	// 			State:        qg.GameStateLobby,
	// 			NextChangeIn: db.ToPgInterval(gm.config.MaxLobbyState),
	// 		})
	// 	})
	// 	if err != nil {
	// 		t.Fatalf("failed to create test game: %v", err)
	// 	}
	// 	upd, err := gm.AdvanceGame(ctx, game.ID)
	// 	if err != nil {
	// 		t.Fatalf("failed to advance test game: %v", err)
	// 	}
	// 	// 2) No time to change state, nPlayer = 0
	// 	require.Nil(t, upd, "update not nil")

	// 	// 3) Add Player 1
	// 	_, err = db.WithTxVQ(ctx, ts.dbManager, func(ctx context.Context, q qg.Querier) (qg.GamePlayer, error) {
	// 		return q.TestAddPlayerToGame(ctx, qg.TestAddPlayerToGameParams{
	// 			GameID:   game.ID,
	// 			PlayerID: players[0].TgID,
	// 		})
	// 	})
	// 	if err != nil {
	// 		t.Fatalf("failed to add player 1 to game: %v", err)
	// 	}

	// 	upd, err = gm.AdvanceGame(ctx, game.ID)
	// 	if err != nil {
	// 		t.Fatalf("failed to advance test game: %v", err)
	// 	}
	// 	// 4) No time to change state, nPlayer = 1
	// 	require.Nil(t, upd, "update not nil")

	// 	// 5) Add Player 2
	// 	_, err = db.WithTxVQ(ctx, ts.dbManager, func(ctx context.Context, q qg.Querier) (qg.GamePlayer, error) {
	// 		return q.TestAddPlayerToGame(ctx, qg.TestAddPlayerToGameParams{
	// 			GameID:   game.ID,
	// 			PlayerID: players[1].TgID,
	// 		})
	// 	})
	// 	if err != nil {
	// 		t.Fatalf("failed to add player 2 to game: %v", err)
	// 	}

	// 	upd, err = gm.AdvanceGame(ctx, game.ID)
	// 	if err != nil {
	// 		t.Fatalf("failed to advance test game: %v", err)
	// 	}
	// 	// 6) No time to change state, nPlayer = 2, but minLobbyTime not exceeded
	// 	require.Nil(t, upd, "update not nil")

	// 	// 7) Fetch game again and check that remaning time reduced
	// 	game, err = db.WithTxVQ(ctx, ts.dbManager, func(ctx context.Context, q qg.Querier) (qg.Game, error) {
	// 		return q.TestGetGameById(ctx, qg.TestGetGameByIdParams{ID: game.ID})
	// 	})
	// 	if err != nil {
	// 		t.Fatalf("failed to fetc game again: %v", err)
	// 	}
	// 	assert.Less(t, game.NextStateChangeAt.Time.Sub(game.StateChangedAt.Time), gm.config.MaxLobbyState)

	// 	// 8) Wait and Advance
	// 	time.Sleep(1 * time.Second)
	// 	upd, err = gm.AdvanceGame(ctx, game.ID)
	// 	if err != nil {
	// 		t.Fatalf("failed to advance test game: %v", err)
	// 	}
	// 	// 9) Game state changed to submitting
	// 	require.NotNil(t, upd)
	// 	assert.Equal(t, qg.GameStateSubmitting, upd.State)
	// 	assert.Equal(t, upd.StateChangeReason, ReasonLobbyFullPlayers)
	// })

	// t.Run("LobbyTimeout", func(t *testing.T) {
	// 	ctx, cancelFunc := context.WithTimeout(context.Background(), 5*time.Second)
	// 	defer cancelFunc()

	// 	// 1) Create game
	// 	game, err := db.WithTxVQ(ctx, ts.dbManager, func(ctx context.Context, q qg.Querier) (qg.Game, error) {
	// 		return q.TestCreateGame(ctx, qg.TestCreateGameParams{
	// 			ThemeID:      theme.ID,
	// 			State:        qg.GameStateLobby,
	// 			NextChangeIn: db.ToPgInterval(gm.config.MaxLobbyState),
	// 		})
	// 	})
	// 	if err != nil {
	// 		t.Fatalf("failed to create test game: %v", err)
	// 	}
	// 	upd, err := gm.AdvanceGame(ctx, game.ID)
	// 	if err != nil {
	// 		t.Fatalf("failed to advance test game: %v", err)
	// 	}
	// 	// 2) No time to change state, nPlayer = 0, no timeout
	// 	require.Nil(t, upd, "update not nil")

	// 	// 3) Wait and Advance
	// 	time.Sleep(2 * time.Second)
	// 	upd, err = gm.AdvanceGame(ctx, game.ID)
	// 	if err != nil {
	// 		t.Fatalf("failed to advance test game: %v", err)
	// 	}
	// 	// 4) Game state changed to submitting
	// 	require.NotNil(t, upd)
	// 	assert.Equal(t, qg.GameStateSubmitting, upd.State)
	// 	assert.Equal(t, upd.StateChangeReason, ReasonLobbyTimeout)
	// })
}
