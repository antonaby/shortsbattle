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

var (
	testUrl1 = "https://www.youtube.com/shorts/GkTZHFyHi1c"
	testUrl2 = "https://www.youtube.com/shorts/zy7xd4zOl7s"
	testUrl3 = "https://www.youtube.com/shorts/wHZyy_G0aYU"
)

func fatalIfError(t *testing.T, err error, msg string) {
	if err != nil {
		t.Fatalf("%s: %v", msg, err)
	}
}

func TestGameActions(t *testing.T) {
	ts := tests.NewDockerTestSuite(t)

	gm := NewGameManager(ts.DBManager, GameConfig{
		MaxPlayers:             2,
		MaxLobbyStage:          2 * time.Second,
		MaxLobbyFullStage:      1 * time.Second,
		MaxSubmitStage:         2 * time.Second,
		MaxSubmitCompleteStage: 1 * time.Second,
		MaxWatchStage:          2 * time.Second,
		MaxWatchCompleteStage:  1 * time.Second,
	})

	vs := NewVideosService(ts.DBManager)

	t.Run("JoinGame", func(t *testing.T) {
		ctx, cancelFunc := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelFunc()

		// create theme and players
		theme, err := tests.CreateTestTheme(ctx, ts.DBManager, 3, qg.GameModeLikedislike)
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

		theme, err := tests.CreateTestTheme(ctx, ts.DBManager, 3, qg.GameModeLikedislike)
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

		videos, err := vs.GetVideosByPlayer(ctx, players[0].TgID, "")
		fatalIfError(t, err, "failed to get videos (player 1)")
		// check that 2 video has been added (as 1 and 2 video are the same)
		require.Equal(t, 2, len(videos))
		require.Equal(t, player1video1.ID, videos[1].ID) // desc order
		require.Equal(t, player1video3.ID, videos[0].ID)

		// create game and add player to it
		game1, err := tests.CreateGameWithStage(ctx, ts.DBManager, theme.ID, qg.GameModeLikedislike, qg.GameStageSubmit, 1)
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
		require.NotEqual(t, player1GameVideo1.Video.ID, player1GameVideo2.Video.ID)

		// add new video
		_, player1GameVideo3, err := gm.SubmitNewVideo(ctx, game1.ID, testUrl3, players[0].TgID, 1)
		fatalIfError(t, err, "failed to submit video (player 1)")
		// check that video has been added as it's new
		require.Equal(t, player1GameVideo1.ID, player1GameVideo3.ID)
		require.NotEqual(t, player1GameVideo1.Video.ID, player1GameVideo3.Video.ID)

		// add new video, but the one that already exist in DB
		_, player1GameVideo4, err := gm.SubmitNewVideo(ctx, game1.ID, testUrl1, players[0].TgID, 1)
		fatalIfError(t, err, "failed to submit video (player 1)")
		// check that video has been updated not added, the video is the same as video 1
		require.Equal(t, player1GameVideo1.ID, player1GameVideo4.ID)
		require.Equal(t, player1video1.ID, player1GameVideo4.Video.ID)

		// add video for player 2
		_, player2GameVideo1, err := gm.SubmitNewVideo(ctx, game1.ID, testUrl1, players[1].TgID, 1)
		fatalIfError(t, err, "failed to submit video (player 2)")
		// should be different game vidoes but the same video
		require.NotEqual(t, player1GameVideo4.ID, player2GameVideo1.ID)
		require.NotEqual(t, player1GameVideo4.PlayerID, player2GameVideo1.PlayerID)
		require.Equal(t, player1GameVideo4.Video.ID, player2GameVideo1.Video.ID)

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
		gamePlayer3, err := tests.ChangePlayerGameMode(ctx, ts.DBManager, players[2].TgID, qg.PlayerGameModeSubmitAndVote)
		if err != nil {
			t.Fatalf("failed to update game mode (player 3): %v", err)
		}
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
		require.Equal(t, player2GameVideo1.Video.ID, videosToWatchP1[0].Video.ID)
		require.Equal(t, player3GameVideo1.Video.ID, videosToWatchP1[1].Video.ID)

		// get videos for player 2
		videosToWatchP2, err := gm.GetVideosToWatch(ctx, game1.ID, players[1].TgID, 1)
		fatalIfError(t, err, "failed to get videos to watch (player 2)")
		// should be videos from player 1 and 3
		require.Equal(t, 2, len(videosToWatchP2))
		require.Equal(t, player1GameVideo4.Video.ID, videosToWatchP2[0].Video.ID)
		require.Equal(t, player3GameVideo1.Video.ID, videosToWatchP2[1].Video.ID)
	})

	t.Run("VoteForVideo", func(t *testing.T) {
		ctx, cancelFunc := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelFunc()

		theme, err := tests.CreateTestTheme(ctx, ts.DBManager, 3, qg.GameModeLikedislike)
		fatalIfError(t, err, "failed to create test theme")

		players, err := tests.CreateTestPlayers(ctx, ts.DBManager, 3, 20)
		fatalIfError(t, err, "failed to create test players")

		game1, err := tests.CreateGameWithStage(ctx, ts.DBManager, theme.ID, qg.GameModeLikedislike, qg.GameStageSubmit, 1)
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

		// player 1 bad
		player1VoteBadData := []byte(`{"value": "abc"}`)
		_, _, rawErr := gm.VoteForVideo(ctx, player2GameVideo1.ID, players[0].TgID, json.RawMessage(player1VoteBadData))
		require.NotNil(t, rawErr)
		badDataErr := rawErr.(common.ServiceError)
		require.Equal(t, common.ErrorBadData, int(badDataErr.Code))

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
		_, _, rawErr = gm.VoteForVideo(ctx, player1GameVideo1.ID, players[0].TgID, json.RawMessage(player1VoteData))
		require.NotNil(t, rawErr)
		forbiddenServiceErr := rawErr.(common.ServiceError)
		require.Equal(t, common.ErrorForbidden, int(forbiddenServiceErr.Code))

		// player can't vote for video if not in the game (player 3)
		_, _, rawErr = gm.VoteForVideo(ctx, player1GameVideo1.ID, players[2].TgID, json.RawMessage(player1VoteData))
		require.NotNil(t, rawErr)
		forbiddenServiceErr = rawErr.(common.ServiceError)
		require.Equal(t, common.ErrorForbidden, int(forbiddenServiceErr.Code))
	})

	t.Run("GetFinalResultRandom", func(t *testing.T) {
		testVideos := []string{
			"https://www.youtube.com/shorts/WYfhopYduI0",
			"https://www.youtube.com/shorts/vjrDd1tg-7U",
			"https://www.youtube.com/shorts/6l_SPhjdKnE",
			"https://www.youtube.com/shorts/B2uK-mAY-GI",
			"https://www.youtube.com/shorts/6re6j4rNyiw",
			"https://www.youtube.com/shorts/2RkJ_UzM_WE",
			"https://www.youtube.com/shorts/rPaIdGOZVHg",
			"https://www.youtube.com/shorts/Mn8bc-TYQQI",
			"https://www.youtube.com/shorts/Wuuo966yQ1k",
			"https://www.youtube.com/shorts/gb5VG9n3eEY",
		}

		ctx, cancelFunc := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelFunc()

		theme, err := tests.CreateTestTheme(ctx, ts.DBManager, 1, qg.GameModeLikedislike)
		if err != nil {
			t.Fatalf("failed to create test theme: %v", err)
		}

		players, err := tests.CreateTestPlayers(ctx, ts.DBManager, 10, 30)
		if err != nil {
			t.Fatalf("failed to create test players: %v", err)
		}

		// 1) create game and add players
		game, err := tests.CreateGameWithStage(ctx, ts.DBManager, theme.ID, qg.GameModeLikedislike, qg.GameStageLobby, 1)
		if err != nil {
			t.Fatalf("failed to create test game: %v", err)
		}
		for _, p := range players {
			_, err = tests.AddPlayerToGame(ctx, ts.DBManager, game.ID, p.TgID, qg.PlayerGameModeSubmitAndVote)
			if err != nil {
				t.Fatalf("failed to add player to game (player %d): %v", p.TgID, err)
			}
		}

		// 2) change game state to submit and add videos
		_, err = tests.ChangeGameStage(ctx, ts.DBManager, game.ID, qg.GameStageSubmit)
		if err != nil {
			t.Fatalf("failed to change game stage: %v", err)
		}
		gameVideos := []models.GameVideo{}
		for i, p := range players {
			url := testVideos[i]
			_, vd, err := gm.SubmitNewVideo(ctx, game.ID, url, p.TgID, 1)
			if err != nil {
				t.Fatalf("failed to submit video (player %d): %v", p.TgID, err)
			}
			gameVideos = append(gameVideos, *vd)
		}

		// 3) change game state to watch and vote
		_, err = tests.ChangeGameStage(ctx, ts.DBManager, game.ID, qg.GameStageWatch)
		if err != nil {
			t.Fatalf("failed to change game stage: %v", err)
		}
		for _, p := range players {
			for i, vd := range gameVideos {
				if vd.PlayerID != p.TgID {
					var vote []byte
					if i%2 == 0 {
						vote = []byte(`{"value": "like"}`)
					} else {
						vote = []byte(`{"value": "dislike"}`)
					}

					_, _, err = gm.VoteForVideo(ctx, vd.ID, p.TgID, vote)
					if err != nil {
						t.Fatalf("failed to vote, p:%d vd:%d : %v", p.TgID, vd.ID, err)
					}
				}
			}
		}

		// 4) change game state to watch complete so it will calculate result
		_, err = tests.ChangeGameStage(ctx, ts.DBManager, game.ID, qg.GameStageWatchComplete)
		if err != nil {
			t.Fatalf("failed to change game stage: %v", err)
		}

		// 5) advacne game and check result
		time.Sleep(1 * time.Second)
		upd, err := gm.AdvanceGameNow(ctx, game.ID)
		if err != nil {
			t.Fatalf("failed to advance test game: %v", err)
		}
		require.NotNil(t, upd)
		require.NotNil(t, upd.Result)
		var finalResult models.LikeDislikeFinalResult
		if err := json.Unmarshal(*upd.Result, &finalResult); err != nil {
			t.Fatalf("failed to unmarchal final result: %v", err)
		}
		require.Equal(t, 10, len(finalResult.Results))

		// 6) get final result for player
		fResult, err := gm.GetFinalResult(ctx, game.ID, players[0].TgID)
		if err != nil {
			t.Fatalf("failed to get final result (player 1): %v", err)
		}

		// it should return the same result
		require.NotNil(t, fResult)
		require.Equal(t, fResult.Result, *upd.Result)

		scores, err := gm.GetPlayerScores(ctx, game.ID, players[0].TgID)
		if err != nil {
			t.Fatalf("failed to get player scores (player 1): %v", err)
		}

		require.Equal(t, 10, len(scores))
		for _, s := range scores {
			if s.PlayerID%2 == 0 {
				assert.Equal(t, int64(9), s.Points)
			} else {
				assert.Equal(t, int64(0), s.Points)
			}
		}
	})

	t.Run("GetFinalResult2Rounds", func(t *testing.T) {
		testVideos := []string{
			"https://www.youtube.com/shorts/WYfhopYduI0",
			"https://www.youtube.com/shorts/vjrDd1tg-7U",
			"https://www.youtube.com/shorts/6l_SPhjdKnE",
			"https://www.youtube.com/shorts/B2uK-mAY-GI",
		}

		ctx, cancelFunc := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelFunc()

		theme, err := tests.CreateTestTheme(ctx, ts.DBManager, 2, qg.GameModeLikedislike)
		if err != nil {
			t.Fatalf("failed to create test theme: %v", err)
		}

		players, err := tests.CreateTestPlayers(ctx, ts.DBManager, 2, 40)
		if err != nil {
			t.Fatalf("failed to create test players: %v", err)
		}

		// 1) create game and add players
		game, err := tests.CreateGameWithStage(ctx, ts.DBManager, theme.ID, qg.GameModeLikedislike, qg.GameStageLobby, 1)
		if err != nil {
			t.Fatalf("failed to create test game: %v", err)
		}
		for _, p := range players {
			_, err = tests.AddPlayerToGame(ctx, ts.DBManager, game.ID, p.TgID, qg.PlayerGameModeSubmitAndVote)
			if err != nil {
				t.Fatalf("failed to add player to game (player %d): %v", p.TgID, err)
			}
		}

		// 2) change game state to submit and add videos
		_, err = tests.ChangeGameStage(ctx, ts.DBManager, game.ID, qg.GameStageSubmit)
		if err != nil {
			t.Fatalf("failed to change game stage: %v", err)
		}
		gameVideos := []models.GameVideo{}
		// round 1
		for i, p := range players {
			url := testVideos[i]
			_, vd, err := gm.SubmitNewVideo(ctx, game.ID, url, p.TgID, 1)
			if err != nil {
				t.Fatalf("failed to submit video (player %d): %v", p.TgID, err)
			}
			gameVideos = append(gameVideos, *vd)
		}
		// round 2
		for i, p := range players {
			url := testVideos[i+2]
			_, vd, err := gm.SubmitNewVideo(ctx, game.ID, url, p.TgID, 2)
			if err != nil {
				t.Fatalf("failed to submit video (player %d): %v", p.TgID, err)
			}
			gameVideos = append(gameVideos, *vd)
		}

		// 3) change game state to watch and vote
		_, err = tests.ChangeGameStage(ctx, ts.DBManager, game.ID, qg.GameStageWatch)
		if err != nil {
			t.Fatalf("failed to change game stage: %v", err)
		}
		// round 1 player 1
		vote := []byte(`{"value": "like"}`)
		_, _, err = gm.VoteForVideo(ctx, gameVideos[1].ID, players[0].TgID, vote)
		if err != nil {
			t.Fatalf("failed to vote for video (player 1): %v", err)
		}
		// round 1 player 2
		vote = []byte(`{"value": "like"}`)
		_, _, err = gm.VoteForVideo(ctx, gameVideos[0].ID, players[1].TgID, vote)
		if err != nil {
			t.Fatalf("failed to vote for video (player 2): %v", err)
		}

		// round 2 player 1
		vote = []byte(`{"value": "dislike"}`)
		_, _, err = gm.VoteForVideo(ctx, gameVideos[3].ID, players[0].TgID, vote)
		if err != nil {
			t.Fatalf("failed to vote for video (player 1): %v", err)
		}
		// round 2 player 2
		vote = []byte(`{"value": "dislike"}`)
		_, _, err = gm.VoteForVideo(ctx, gameVideos[2].ID, players[1].TgID, vote)
		if err != nil {
			t.Fatalf("failed to vote for video (player 2): %v", err)
		}

		// 4) change game state to watch complete so it will calculate result
		_, err = tests.ChangeGameStageWithRound(ctx, ts.DBManager, game.ID, qg.GameStageWatchComplete, 2)
		if err != nil {
			t.Fatalf("failed to change game stage: %v", err)
		}

		// 5) advacne game and check result
		time.Sleep(1 * time.Second)
		upd, err := gm.AdvanceGameNow(ctx, game.ID)
		if err != nil {
			t.Fatalf("failed to advance test game: %v", err)
		}
		require.NotNil(t, upd)
		require.NotNil(t, upd.Result)

		scores, err := gm.GetPlayerScores(ctx, game.ID, players[0].TgID)
		if err != nil {
			t.Fatalf("failed to get player scores (player 1): %v", err)
		}

		require.Equal(t, 2, len(scores))
		require.Equal(t, int64(1), scores[0].Points)
		require.Equal(t, int64(1), scores[1].Points)
	})
}

func TestAdvanceGame(t *testing.T) {
	ts := tests.NewDockerTestSuite(t)

	gm := NewGameManager(ts.DBManager, GameConfig{
		MaxPlayers:             2,
		MaxLobbyStage:          2 * time.Second,
		MaxLobbyFullStage:      1 * time.Second,
		MaxSubmitStage:         2 * time.Second,
		MaxSubmitCompleteStage: 1 * time.Second,
		MaxWatchStage:          2 * time.Second,
		MaxWatchCompleteStage:  1 * time.Second,
	})

	t.Run("LobbyWithMaxPlayers", func(t *testing.T) {
		ctx, cancelFunc := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelFunc()

		theme, err := tests.CreateTestTheme(ctx, ts.DBManager, 3, qg.GameModeLikedislike)
		if err != nil {
			t.Fatalf("failed to create test theme: %v", err)
		}

		players, err := tests.CreateTestPlayers(ctx, ts.DBManager, 2, 0)
		if err != nil {
			t.Fatalf("failed to create test players: %v", err)
		}

		// 1) Create game
		game, err := tests.CreateGameWithStage(ctx, ts.DBManager, theme.ID, qg.GameModeLikedislike, qg.GameStageLobby, 0)
		if err != nil {
			t.Fatalf("failed to create test game: %v", err)
		}
		upd, err := gm.AdvanceGameNow(ctx, game.ID)
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
		upd, err = gm.AdvanceGameNow(ctx, game.ID)
		if err != nil {
			t.Fatalf("failed to advance test game: %v", err)
		}
		// 4) No time to change state, nPlayer = 1
		require.Nil(t, upd, "update not nil")

		// 5) Add Player 2
		_, err = tests.AddPlayerToGame(ctx, ts.DBManager, game.ID, players[1].TgID, qg.PlayerGameModeSubmitAndVote)
		if err != nil {
			t.Fatalf("failed to add player to game (player 2): %v", err)
		}
		upd, err = gm.AdvanceGameNow(ctx, game.ID)
		if err != nil {
			t.Fatalf("failed to advance test game: %v", err)
		}

		// 6) Game state changed to lobby-full
		require.NotNil(t, upd)
		require.Equal(t, qg.GameStageLobbyFull, upd.Stage)
		require.Equal(t, *upd.StateChangeReason, models.ReasonLobbyFull)
	})

	t.Run("LobbyTimeout", func(t *testing.T) {
		ctx, cancelFunc := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelFunc()

		theme, err := tests.CreateTestTheme(ctx, ts.DBManager, 3, qg.GameModeLikedislike)
		if err != nil {
			t.Fatalf("failed to create test theme: %v", err)
		}

		// 1) Create game
		game, err := tests.CreateGameWithStage(ctx, ts.DBManager, theme.ID, qg.GameModeLikedislike, qg.GameStageLobby, 0)
		if err != nil {
			t.Fatalf("failed to create test game: %v", err)
		}
		upd, err := gm.AdvanceGameNow(ctx, game.ID)
		if err != nil {
			t.Fatalf("failed to advance test game: %v", err)
		}
		// 2) No time to change state, nPlayer = 0
		require.Nil(t, upd, "update not nil")

		// 3) sleep for 2 seconds and advance game
		time.Sleep(2 * time.Second)
		upd, err = gm.AdvanceGameNow(ctx, game.ID)
		if err != nil {
			t.Fatalf("failed to advance test game: %v", err)
		}

		// 4) Game state changed to lobby-full because of timeout
		require.NotNil(t, upd)
		require.Equal(t, qg.GameStageLobbyFull, upd.Stage)
		require.Equal(t, *upd.StateChangeReason, models.ReasonLobbyTimeout)
		// 5) should be round 0 as game not started
		require.Equal(t, upd.RoundN, int32(0))
	})

	t.Run("LobbyFullTimeout", func(t *testing.T) {
		ctx, cancelFunc := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelFunc()

		theme, err := tests.CreateTestTheme(ctx, ts.DBManager, 3, qg.GameModeLikedislike)
		if err != nil {
			t.Fatalf("failed to create test theme: %v", err)
		}

		// 1) Create game
		game, err := tests.CreateGameWithStage(ctx, ts.DBManager, theme.ID, qg.GameModeLikedislike, qg.GameStageLobbyFull, 0)
		if err != nil {
			t.Fatalf("failed to create test game: %v", err)
		}
		upd, err := gm.AdvanceGameNow(ctx, game.ID)
		if err != nil {
			t.Fatalf("failed to advance test game: %v", err)
		}
		// 2) No time to change state, no time
		require.Nil(t, upd, "update not nil")

		// 3) sleep for 1 seconds and advance game
		time.Sleep(1 * time.Second)
		upd, err = gm.AdvanceGameNow(ctx, game.ID)
		if err != nil {
			t.Fatalf("failed to advance test game: %v", err)
		}

		// 4) Game state changed to submit because of timeout
		require.NotNil(t, upd)
		require.Equal(t, qg.GameStageSubmit, upd.Stage)
		require.Equal(t, *upd.StateChangeReason, models.ReasonLobbyFullTimeout)
		// 5) should be round 1 as game started
		require.Equal(t, upd.RoundN, int32(1))
	})

	t.Run("SubmitAll", func(t *testing.T) {
		ctx, cancelFunc := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelFunc()

		theme, err := tests.CreateTestTheme(ctx, ts.DBManager, 3, qg.GameModeLikedislike)
		if err != nil {
			t.Fatalf("failed to create test theme: %v", err)
		}

		players, err := tests.CreateTestPlayers(ctx, ts.DBManager, 2, 10)
		if err != nil {
			t.Fatalf("failed to create test players: %v", err)
		}

		// 1) Create game
		game, err := tests.CreateGameWithStage(ctx, ts.DBManager, theme.ID, qg.GameModeLikedislike, qg.GameStageSubmit, 1)
		if err != nil {
			t.Fatalf("failed to create test game: %v", err)
		}

		// 2) add player to game
		_, err = tests.AddPlayerToGame(ctx, ts.DBManager, game.ID, players[0].TgID, qg.PlayerGameModeSubmitAndVote)
		if err != nil {
			t.Fatalf("failed to add player to game (player 1): %v", err)
		}
		_, err = tests.AddPlayerToGame(ctx, ts.DBManager, game.ID, players[1].TgID, qg.PlayerGameModeSubmitAndVote)
		if err != nil {
			t.Fatalf("failed to add player to game (player 2): %v", err)
		}

		// 3) submit video for player 1
		_, _, err = gm.SubmitNewVideo(ctx, game.ID, testUrl1, players[0].TgID, 1)
		if err != nil {
			t.Fatalf("failed to submit video (player 1): %v", err)
		}

		// 4) advance game
		upd, err := gm.AdvanceGameNow(ctx, game.ID)
		if err != nil {
			t.Fatalf("failed to advance test game: %v", err)
		}
		// No time to change state, only one player submitted video
		require.Nil(t, upd, "update not nil")

		// 5) submit video for player 2
		_, _, err = gm.SubmitNewVideo(ctx, game.ID, testUrl2, players[1].TgID, 1)
		if err != nil {
			t.Fatalf("failed to submit video (player 2): %v", err)
		}

		// 6) advance game
		upd, err = gm.AdvanceGameNow(ctx, game.ID)
		if err != nil {
			t.Fatalf("failed to advance test game: %v", err)
		}
		// No time to change state, only one player submitted video
		require.NotNil(t, upd, "update nil")
		require.Equal(t, qg.GameStageSubmitComplete, upd.Stage)
		require.Equal(t, *upd.StateChangeReason, models.ReasonSubmitAll)
	})

	t.Run("SubmitTimeout", func(t *testing.T) {
		ctx, cancelFunc := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelFunc()

		theme, err := tests.CreateTestTheme(ctx, ts.DBManager, 3, qg.GameModeLikedislike)
		if err != nil {
			t.Fatalf("failed to create test theme: %v", err)
		}

		// 1) Create game
		game, err := tests.CreateGameWithStage(ctx, ts.DBManager, theme.ID, qg.GameModeLikedislike, qg.GameStageSubmit, 1)
		if err != nil {
			t.Fatalf("failed to create test game: %v", err)
		}

		// 2) advance game
		upd, err := gm.AdvanceGameNow(ctx, game.ID)
		if err != nil {
			t.Fatalf("failed to advance test game: %v", err)
		}
		// No time to change state, no videos
		require.Nil(t, upd, "update not nil")

		time.Sleep(2 * time.Second)
		// 3) advance game after timeout
		upd, err = gm.AdvanceGameNow(ctx, game.ID)
		if err != nil {
			t.Fatalf("failed to advance test game: %v", err)
		}

		// 4) Game state changed to submit-complete because of timeout
		require.NotNil(t, upd)
		require.Equal(t, qg.GameStageSubmitComplete, upd.Stage)
		require.Equal(t, *upd.StateChangeReason, models.ReasonSubmitTimeout)
	})

	t.Run("SubmitCompleteTimeout", func(t *testing.T) {
		ctx, cancelFunc := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelFunc()

		theme, err := tests.CreateTestTheme(ctx, ts.DBManager, 3, qg.GameModeLikedislike)
		if err != nil {
			t.Fatalf("failed to create test theme: %v", err)
		}

		// 1) Create game
		game, err := tests.CreateGameWithStage(ctx, ts.DBManager, theme.ID, qg.GameModeLikedislike, qg.GameStageSubmitComplete, 1)
		if err != nil {
			t.Fatalf("failed to create test game: %v", err)
		}
		upd, err := gm.AdvanceGameNow(ctx, game.ID)
		if err != nil {
			t.Fatalf("failed to advance test game: %v", err)
		}
		// 2) No time to change state, no time
		require.Nil(t, upd, "update not nil")

		// 3) sleep for 1 seconds and advance game
		time.Sleep(1 * time.Second)
		upd, err = gm.AdvanceGameNow(ctx, game.ID)
		if err != nil {
			t.Fatalf("failed to advance test game: %v", err)
		}

		// 4) Game state changed to submit because of timeout
		require.NotNil(t, upd)
		require.Equal(t, qg.GameStageWatch, upd.Stage)
		require.Equal(t, *upd.StateChangeReason, models.ReasinSubmitCompleteTimeout)
		// 5) should be round 1 as game started
		require.Equal(t, upd.RoundN, int32(1))
	})

	t.Run("WatchAll", func(t *testing.T) {
		ctx, cancelFunc := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelFunc()

		theme, err := tests.CreateTestTheme(ctx, ts.DBManager, 3, qg.GameModeLikedislike)
		if err != nil {
			t.Fatalf("failed to create test theme: %v", err)
		}
		players, err := tests.CreateTestPlayers(ctx, ts.DBManager, 2, 20)
		if err != nil {
			t.Fatalf("failed to create test players: %v", err)
		}

		// 1) Create game
		game, err := tests.CreateGameWithStage(ctx, ts.DBManager, theme.ID, qg.GameModeLikedislike, qg.GameStageSubmit, 1)
		if err != nil {
			t.Fatalf("failed to create test game: %v", err)
		}

		// 2) add players to game
		_, err = tests.AddPlayerToGame(ctx, ts.DBManager, game.ID, players[0].TgID, qg.PlayerGameModeSubmitAndVote)
		if err != nil {
			t.Fatalf("failed to add player to game (player 1): %v", err)
		}
		_, err = tests.AddPlayerToGame(ctx, ts.DBManager, game.ID, players[1].TgID, qg.PlayerGameModeSubmitAndVote)
		if err != nil {
			t.Fatalf("failed to add player to game (player 2): %v", err)
		}

		// 3) submit videos for players
		_, vd1, err := gm.SubmitNewVideo(ctx, game.ID, testUrl1, players[0].TgID, 1)
		if err != nil {
			t.Fatalf("failed to submit video (player 1): %v", err)
		}
		_, vd2, err := gm.SubmitNewVideo(ctx, game.ID, testUrl2, players[1].TgID, 1)
		if err != nil {
			t.Fatalf("failed to submit video (player 2): %v", err)
		}

		// 4) chage game stage
		_, err = tests.ChangeGameStage(ctx, ts.DBManager, game.ID, qg.GameStageWatch)
		if err != nil {
			t.Fatalf("failed to change game stage: %v", err)
		}

		// 5) player 1 votes
		player1Vote := []byte(`{"value": "like"}`)
		_, _, err = gm.VoteForVideo(ctx, vd2.ID, players[0].TgID, player1Vote)
		if err != nil {
			t.Fatalf("failed to vote for video (player 1): %v", err)
		}

		// 6) advance game
		upd, err := gm.AdvanceGameNow(ctx, game.ID)
		if err != nil {
			t.Fatalf("failed to advance test game: %v", err)
		}
		// no update as only 1 player voted
		require.Nil(t, upd)

		// 5) player 1 votes
		player2Vote := []byte(`{"value": "like"}`)
		_, _, err = gm.VoteForVideo(ctx, vd1.ID, players[1].TgID, player2Vote)
		if err != nil {
			t.Fatalf("failed to vote for video (player 2): %v", err)
		}

		// 6) advance game
		upd, err = gm.AdvanceGameNow(ctx, game.ID)
		if err != nil {
			t.Fatalf("failed to advance test game: %v", err)
		}
		// updated as both players voted
		require.NotNil(t, upd)
		require.Equal(t, qg.GameStageWatchComplete, upd.Stage)
		require.Equal(t, *upd.StateChangeReason, models.ReasonWatchAll)
	})

	t.Run("WatchTimeout", func(t *testing.T) {
		ctx, cancelFunc := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelFunc()

		theme, err := tests.CreateTestTheme(ctx, ts.DBManager, 3, qg.GameModeLikedislike)
		if err != nil {
			t.Fatalf("failed to create test theme: %v", err)
		}

		// 1) Create game
		game, err := tests.CreateGameWithStage(ctx, ts.DBManager, theme.ID, qg.GameModeLikedislike, qg.GameStageWatch, 1)
		if err != nil {
			t.Fatalf("failed to create test game: %v", err)
		}

		// 2) advance game
		upd, err := gm.AdvanceGameNow(ctx, game.ID)
		if err != nil {
			t.Fatalf("failed to advance test game: %v", err)
		}
		// No time to change state, no votes
		require.Nil(t, upd, "update not nil")

		time.Sleep(2 * time.Second)
		// 3) advance game after timeout
		upd, err = gm.AdvanceGameNow(ctx, game.ID)
		if err != nil {
			t.Fatalf("failed to advance test game: %v", err)
		}

		// 4) Game state changed to submit-complete because of timeout
		require.NotNil(t, upd)
		require.Equal(t, qg.GameStageWatchComplete, upd.Stage)
		require.Equal(t, *upd.StateChangeReason, models.ReasonWatchTimeout)
	})

	t.Run("WatchComplete", func(t *testing.T) {
		ctx, cancelFunc := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelFunc()

		theme, err := tests.CreateTestTheme(ctx, ts.DBManager, 2, qg.GameModeLikedislike)
		if err != nil {
			t.Fatalf("failed to create test theme: %v", err)
		}

		// 1) Create game
		game, err := tests.CreateGameWithStage(ctx, ts.DBManager, theme.ID, qg.GameModeLikedislike, qg.GameStageWatchComplete, 1)
		if err != nil {
			t.Fatalf("failed to create test game: %v", err)
		}
		// 2) advance game
		upd, err := gm.AdvanceGameNow(ctx, game.ID)
		if err != nil {
			t.Fatalf("failed to advance test game: %v", err)
		}
		// No time to change state, no timeout
		require.Nil(t, upd, "update not nil")

		// 3) wait for the next round
		time.Sleep(1 * time.Second)
		upd, err = gm.AdvanceGameNow(ctx, game.ID)
		if err != nil {
			t.Fatalf("failed to advance test game: %v", err)
		}
		// No time to change state, no timeout
		require.NotNil(t, upd, "update not nil")
		require.Equal(t, qg.GameStageSubmit, upd.Stage)
		require.Equal(t, *upd.StateChangeReason, models.ReasonWatchCompleteNextRound)
		// should be second round
		require.Equal(t, upd.RoundN, int32(2))

		// 4) chnage game state back to watch complete and advance
		_, err = tests.ChangeGameStageAndUpdateAt(ctx, ts.DBManager, game.ID, qg.GameStageWatchComplete, gm.config.MaxWatchCompleteStage)
		if err != nil {
			t.Fatalf("failed to change game state: %v", err)
		}
		time.Sleep(1 * time.Second)
		upd, err = gm.AdvanceGameNow(ctx, game.ID)
		if err != nil {
			t.Fatalf("failed to advance test game: %v", err)
		}
		// game complete
		require.NotNil(t, upd, "update not nil")
		require.Equal(t, qg.GameStageComplete, upd.Stage)
		require.Equal(t, *upd.StateChangeReason, models.ReasonWatchCompleteGameComplete)
		// round is 0 as game complete
		require.Equal(t, upd.RoundN, int32(0))
	})
}
