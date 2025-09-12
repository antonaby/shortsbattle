package services

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/antonaby/shortsbattle/game-server/internal/common"
	"github.com/antonaby/shortsbattle/game-server/internal/db/qg"
	"github.com/antonaby/shortsbattle/game-server/internal/models"
	"github.com/antonaby/shortsbattle/game-server/internal/tests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func fatalIfError(t *testing.T, err error, msg string) {
	if err != nil {
		t.Fatalf("%s: %v", msg, err)
	}
}

func TestGameActions(t *testing.T) {
	testUrl1 := "https://www.youtube.com/shorts/GkTZHFyHi1c"
	testUrl2 := "https://www.youtube.com/shorts/zy7xd4zOl7s"
	testUrl3 := "https://www.youtube.com/shorts/wHZyy_G0aYU"

	ts := tests.NewDockerTestSuite(t)

	gm := NewGameManager(ts.DBManager, GameConfig{
		MaxPlayers:     2,
		MaxLobbyStage:  2 * time.Second,
		MaxSubmitState: 60 * time.Second,
		MaxWatchState:  600 * time.Second,
	})

	vs := NewVideosService(ts.DBManager)

	t.Run("JoinGame", func(t *testing.T) {
		ctx, cancelFunc := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelFunc()

		// create theme and players
		theme, err := tests.CreateTestTheme(ctx, ts.DBManager, 3)
		fatalIfError(t, err, "failed to create test theme")
		players, err := tests.CreateTestPlayers(ctx, ts.DBManager, 3, 0)
		fatalIfError(t, err, "failed to create test players")

		// player 1 joins the game
		gameIdPlayer1Attempt1, err := gm.JoinGame(ctx, theme.ID, players[0].TgID, qg.PlayerGameModeSubmitAndVote)
		fatalIfError(t, err, "failed to join game (player 1)")
		// player 1 should join the same game (as it's already in)
		gameIdPlayer1Attempt2, err := gm.JoinGame(ctx, theme.ID, players[0].TgID, qg.PlayerGameModeSubmitAndVote)
		fatalIfError(t, err, "failed to join game again (player 1)")
		// check that the game is the same
		require.Equal(t, gameIdPlayer1Attempt1, gameIdPlayer1Attempt2)

		// player 2 should join the same game
		gameIdPlayer2Attempt1, err := gm.JoinGame(ctx, theme.ID, players[1].TgID, qg.PlayerGameModeSubmitAndVote)
		fatalIfError(t, err, "failed to join game (player 2)")
		// check that the game is the same
		require.Equal(t, gameIdPlayer1Attempt2, gameIdPlayer2Attempt1)

		// player 3 should join new game (MaxPlayers = 2 therefore it should create a new game)
		gameIdPlayer3Attempt1, err := gm.JoinGame(ctx, theme.ID, players[2].TgID, qg.PlayerGameModeSubmitAndVote)
		fatalIfError(t, err, "failed to join game (player 3)")
		// should be new game as lobby closed (LobbyClosedBefore)
		require.NotEqual(t, gameIdPlayer2Attempt1, gameIdPlayer3Attempt1)

		// get game status for game 1 (player 1)
		gameDetails, err := gm.GetGameDetailsForPlayer(ctx, gameIdPlayer1Attempt2, players[0].TgID)
		fatalIfError(t, err, "failed to get game details (player 1)")
		require.Equal(t, qg.GameStageLobby, gameDetails.Stage)

		// player can't get game details for game it's not joined in (player 1 -> game 2)
		_, rawErr := gm.GetGameDetailsForPlayer(ctx, gameIdPlayer3Attempt1, players[0].TgID)
		require.NotNil(t, rawErr)
		forbiddenErr := rawErr.(common.ServiceError)
		require.Equal(t, common.ErrorForbidden, int(forbiddenErr.Code))

		// can't join game with unknown theme
		_, rawErr = gm.JoinGame(ctx, 10, players[0].TgID, qg.PlayerGameModeSubmitAndVote)
		require.NotNil(t, rawErr)
		notFoundErr := rawErr.(common.ServiceError)
		require.Equal(t, common.ErrorNotFound, int(notFoundErr.Code))
	})

	t.Run("SubmitVideo", func(t *testing.T) {
		ctx, cancelFunc := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelFunc()

		theme, err := tests.CreateTestTheme(ctx, ts.DBManager, 3)
		fatalIfError(t, err, "failed to create test theme")

		players, err := tests.CreateTestPlayers(ctx, ts.DBManager, 4, 10)
		fatalIfError(t, err, "failed to create test players")

		// add 3 videos to player 1
		player1video1, err := vs.AddVideo(ctx, players[0].TgID, testUrl1)
		fatalIfError(t, err, "failed to add video (player 1)")
		// add the same url
		player1video2, err := vs.AddVideo(ctx, players[0].TgID, testUrl1)
		fatalIfError(t, err, "failed to add video (player 1)")
		// so it should return the same video
		require.Equal(t, player1video1.ID, player1video2.ID)
		player1video3, err := vs.AddVideo(ctx, players[0].TgID, testUrl2)
		fatalIfError(t, err, "failed to add video (player 1)")

		videos, err := vs.GetVideosByPlayer(ctx, players[0].TgID)
		fatalIfError(t, err, "failed to get videos (player 1)")
		// check that 2 video has been added (as 1 and 2 video are the same)
		require.Equal(t, 2, len(videos))
		require.Equal(t, player1video1.ID, videos[0].ID)
		require.Equal(t, player1video3.ID, videos[1].ID)

		// create game and add player to it
		game1, err := tests.CreateGameWithStage(ctx, ts.DBManager, theme.ID, qg.GameStageSubmit)
		fatalIfError(t, err, "failed to create test game")
		_, err = tests.AddPlayerToGame(ctx, ts.DBManager, game1.ID, players[0].TgID, qg.PlayerGameModeSubmitAndVote)
		fatalIfError(t, err, "failed to add player to game (player 1)")
		_, err = tests.AddPlayerToGame(ctx, ts.DBManager, game1.ID, players[1].TgID, qg.PlayerGameModeSubmitAndVote)
		fatalIfError(t, err, "failed to add player to game (player 2)")
		_, err = tests.AddPlayerToGame(ctx, ts.DBManager, game1.ID, players[2].TgID, qg.PlayerGameModeOnlyVote)
		fatalIfError(t, err, "failed to add player to game (player 3)")

		// add existing video 1
		_, player1GameVideo1, err := gm.SubmitExistingVideo(ctx, game1.ID, player1video1.ID, players[0].TgID, 1)
		fatalIfError(t, err, "failed to submit video (player 1)")
		// check that video has been added to the same game and player
		require.Equal(t, game1.ID, player1GameVideo1.GameID)
		require.Equal(t, players[0].TgID, player1GameVideo1.PlayerID)

		// add existing video 3
		_, player1GameVideo2, err := gm.SubmitExistingVideo(ctx, game1.ID, player1video3.ID, players[0].TgID, 1)
		fatalIfError(t, err, "failed to submit video (player 1)")
		// check that video has been updated not added
		require.Equal(t, player1GameVideo1.ID, player1GameVideo2.ID)
		require.NotEqual(t, player1GameVideo1.VideoID, player1GameVideo2.VideoID)

		// add new video
		_, player1GameVideo3, err := gm.SubmitNewVideo(ctx, game1.ID, testUrl3, players[0].TgID, 1)
		fatalIfError(t, err, "failed to submit video (player 1)")
		// check that video has been added as it's new
		require.Equal(t, player1GameVideo1.ID, player1GameVideo3.ID)
		require.NotEqual(t, player1GameVideo1.VideoID, player1GameVideo3.VideoID)

		// add new video, but the one that already exist in DB
		_, player1GameVideo4, err := gm.SubmitNewVideo(ctx, game1.ID, testUrl1, players[0].TgID, 1)
		fatalIfError(t, err, "failed to submit video (player 1)")
		// check that video has been updated not added, the video is the same as video 1
		require.Equal(t, player1GameVideo1.ID, player1GameVideo4.ID)
		require.Equal(t, player1video1.ID, player1GameVideo4.VideoID)

		// add video for player 2
		_, player2GameVideo1, err := gm.SubmitNewVideo(ctx, game1.ID, testUrl1, players[1].TgID, 1)
		fatalIfError(t, err, "failed to submit video (player 2)")
		// should be different game vidoes but the same video
		require.NotEqual(t, player1GameVideo4.ID, player2GameVideo1.ID)
		require.NotEqual(t, player1GameVideo4.PlayerID, player2GameVideo1.PlayerID)
		require.Equal(t, player1GameVideo4.VideoID, player2GameVideo1.VideoID)

		// player 2 can't submit video owned by player 1
		_, _, rawErr := gm.SubmitExistingVideo(ctx, game1.ID, player1video3.ID, players[1].TgID, 1)
		require.NotNil(t, rawErr)
		notFoundErr := rawErr.(common.ServiceError)
		require.Equal(t, common.ErrorNotFound, int(notFoundErr.Code))

		// player 3 can't add videos as it's in only watching mode
		_, _, rawErr = gm.SubmitNewVideo(ctx, game1.ID, testUrl1, players[2].TgID, 1)
		require.NotNil(t, rawErr)
		wrongModeErr := rawErr.(common.ServiceError)
		require.Equal(t, common.ErrorForbidden, int(wrongModeErr.Code))

		// chnage game mode for player 3
		gamePlayer3, err := gm.UpdateGameMode(ctx, game1.ID, players[2].TgID, qg.PlayerGameModeSubmitAndVote)
		fatalIfError(t, err, "failed to update game mode (player 3)")
		require.Equal(t, qg.PlayerGameModeSubmitAndVote, gamePlayer3.Mode)

		// add video for player 3
		_, player3GameVideo1, err := gm.SubmitNewVideo(ctx, game1.ID, testUrl3, players[2].TgID, 1)
		fatalIfError(t, err, "failed to submit video (player 3)")
		// should be added to the game
		require.Equal(t, game1.ID, player3GameVideo1.GameID)

		// player 4 not in the game, so it can't submit video
		_, _, rawErr = gm.SubmitNewVideo(ctx, game1.ID, testUrl1, players[3].TgID, 1)
		require.NotNil(t, rawErr)
		forbiddenErr := rawErr.(common.ServiceError)
		require.Equal(t, common.ErrorForbidden, int(forbiddenErr.Code))

		// chnage game stage to watch to be able to get videos to watch
		_, err = tests.ChangeGameStage(ctx, ts.DBManager, game1.ID, qg.GameStageWatch)
		fatalIfError(t, err, "failed to chnage game stage")

		// can't submit in the wrong game stage
		_, _, rawErr = gm.SubmitExistingVideo(ctx, game1.ID, player1video3.ID, players[0].TgID, 1)
		require.NotNil(t, rawErr)
		wrongGameSatgeErr := rawErr.(common.ServiceError)
		require.Equal(t, common.ErrorForbidden, int(wrongGameSatgeErr.Code))

		// get videos for player 1
		videosToWatchP1, err := gm.GetVideosToWatch(ctx, game1.ID, players[0].TgID, 1)
		fatalIfError(t, err, "failed to get videos to watch (player 1)")
		// should be videos from player 2 and 3
		require.Equal(t, 2, len(videosToWatchP1))
		require.Equal(t, player2GameVideo1.VideoID, videosToWatchP1[0].VideoID)
		require.Equal(t, player3GameVideo1.VideoID, videosToWatchP1[1].VideoID)

		// get videos for player 2
		videosToWatchP2, err := gm.GetVideosToWatch(ctx, game1.ID, players[1].TgID, 1)
		fatalIfError(t, err, "failed to get videos to watch (player 2)")
		// should be videos from player 1 and 3
		require.Equal(t, 2, len(videosToWatchP2))
		require.Equal(t, player1GameVideo4.VideoID, videosToWatchP2[0].VideoID)
		require.Equal(t, player3GameVideo1.VideoID, videosToWatchP2[1].VideoID)
	})

	t.Run("VoteForVideo", func(t *testing.T) {
		ctx, cancelFunc := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelFunc()

		theme, err := tests.CreateTestTheme(ctx, ts.DBManager, 3)
		fatalIfError(t, err, "failed to create test theme")

		players, err := tests.CreateTestPlayers(ctx, ts.DBManager, 3, 20)
		fatalIfError(t, err, "failed to create test players")

		game1, err := tests.CreateGameWithStage(ctx, ts.DBManager, theme.ID, qg.GameStageSubmit)
		fatalIfError(t, err, "failed to create test game")

		// add players to the game
		_, err = tests.AddPlayerToGame(ctx, ts.DBManager, game1.ID, players[0].TgID, qg.PlayerGameModeSubmitAndVote)
		fatalIfError(t, err, "failed to add player to game (player 1)")
		_, err = tests.AddPlayerToGame(ctx, ts.DBManager, game1.ID, players[1].TgID, qg.PlayerGameModeSubmitAndVote)
		fatalIfError(t, err, "failed to add player to game (player 2)")

		// add new video (player 1)
		_, player1GameVideo1, err := gm.SubmitNewVideo(ctx, game1.ID, testUrl1, players[0].TgID, 1)
		fatalIfError(t, err, "failed to submit video (player 1)")
		// add new video (player 2)
		_, player2GameVideo1, err := gm.SubmitNewVideo(ctx, game1.ID, testUrl2, players[1].TgID, 1)
		fatalIfError(t, err, "failed to submit video (player 2)")

		// chnage game state to watch
		_, err = tests.ChangeGameStage(ctx, ts.DBManager, game1.ID, qg.GameStageWatch)
		fatalIfError(t, err, "failed to chnage game stage")

		// player 1 vote
		player1VoteData := []byte(`{"value": "like"}`)
		_, player1Vote1, err := gm.VoteForVideo(ctx, player2GameVideo1.ID, players[0].TgID, json.RawMessage(player1VoteData))
		fatalIfError(t, err, "failed to vote (player 1)")
		require.Equal(t, player2GameVideo1.ID, player1Vote1.GameVideoID)

		// player 2 vote
		player2VoteData := []byte(`{"value": "like"}`)
		_, player1Vote2, err := gm.VoteForVideo(ctx, player1GameVideo1.ID, players[1].TgID, json.RawMessage(player2VoteData))
		fatalIfError(t, err, "failed to vote (player 2)")
		require.Equal(t, player1GameVideo1.ID, player1Vote2.GameVideoID)

		// player can't vote for it's own video (palyer 1)
		_, _, rawErr := gm.VoteForVideo(ctx, player1GameVideo1.ID, players[0].TgID, json.RawMessage(player1VoteData))
		require.NotNil(t, rawErr)
		forbiddenServiceErr := rawErr.(common.ServiceError)
		require.Equal(t, common.ErrorForbidden, int(forbiddenServiceErr.Code))

		// player can't vote for video if not in the game (player 3)
		_, _, rawErr = gm.VoteForVideo(ctx, player1GameVideo1.ID, players[2].TgID, json.RawMessage(player1VoteData))
		require.NotNil(t, rawErr)
		forbiddenServiceErr = rawErr.(common.ServiceError)
		require.Equal(t, common.ErrorForbidden, int(forbiddenServiceErr.Code))
	})
}

func TestAdvanceGame(t *testing.T) {
	ts := tests.NewDockerTestSuite(t)

	gm := NewGameManager(ts.DBManager, GameConfig{
		MaxPlayers:     2,
		MaxLobbyStage:  2 * time.Second,
		MaxSubmitState: 60 * time.Second,
		MaxWatchState:  600 * time.Second,
	})

	t.Run("LobbyWithMaxPlayers", func(t *testing.T) {
		ctx, cancelFunc := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelFunc()

		theme, err := tests.CreateTestTheme(ctx, ts.DBManager, 3)
		if err != nil {
			t.Fatalf("failed to create test theme: %v", err)
		}

		players, err := tests.CreateTestPlayers(ctx, ts.DBManager, 2, 0)
		if err != nil {
			t.Fatalf("failed to create test theme: %v", err)
		}

		// 1) Create game
		game, err := tests.CreateGameWithStage(ctx, ts.DBManager, theme.ID, qg.GameStageLobby)
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
		_, err = tests.AddPlayerToGame(ctx, ts.DBManager, game.ID, players[0].TgID, qg.PlayerGameModeSubmitAndVote)
		if err != nil {
			t.Fatalf("failed to add player to game (player 1): %v", err)
		}
		upd, err = gm.AdvanceGame(ctx, game.ID)
		if err != nil {
			t.Fatalf("failed to advance test game: %v", err)
		}
		// 4) No time to change state, nPlayer = 1
		require.Nil(t, upd, "update not nil")

		// 5) Add Player 2
		_, err = tests.AddPlayerToGame(ctx, ts.DBManager, game.ID, players[1].TgID, qg.PlayerGameModeSubmitAndVote)
		if err != nil {
			t.Fatalf("failed to add player to game (player 1): %v", err)
		}
		upd, err = gm.AdvanceGame(ctx, game.ID)
		if err != nil {
			t.Fatalf("failed to advance test game: %v", err)
		}
		// 6) No time to change state, nPlayer = 2, but minLobbyTime not exceeded
		require.Nil(t, upd, "update not nil")

		// 8) Wait and Advance
		time.Sleep(1 * time.Second)
		upd, err = gm.AdvanceGame(ctx, game.ID)
		if err != nil {
			t.Fatalf("failed to advance test game: %v", err)
		}

		// 9) Game state changed to submitting
		require.NotNil(t, upd)
		assert.Equal(t, qg.GameStageSubmit, upd.Stage)
		assert.Equal(t, upd.StateChangeReason, models.ReasonLobbyFull)
	})

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
