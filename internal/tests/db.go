package tests

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/antonaby/shortsbattle/game-server/internal/db"
	"github.com/antonaby/shortsbattle/game-server/internal/db/qg"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
)

func CreateDockerPool() (*dockertest.Pool, error) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		return nil, fmt.Errorf("failed to create pool: %w", err)
	}

	err = pool.Client.Ping()
	if err != nil {
		return nil, fmt.Errorf("failed to ping pool: %w", err)
	}

	return pool, nil
}

func CreatePostgresContainer(t *testing.T, pool *dockertest.Pool, pgUser, pgPassword, pgDb string) (*dockertest.Resource, error) {
	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "17",
		Env: []string{
			fmt.Sprintf("POSTGRES_USER=%s", pgUser),
			fmt.Sprintf("POSTGRES_PASSWORD=%s", pgPassword),
			fmt.Sprintf("POSTGRES_DB=%s", pgDb),
		},
	}, func(config *docker.HostConfig) {
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{
			Name: "no",
		}
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create pg container: %w", err)
	}

	t.Cleanup(func() {
		if err := pool.Purge(resource); err != nil {
			log.Printf("failed to purge docker resource: %v", err)
		}
	})

	return resource, nil
}

func CreateDbManager(pool *dockertest.Pool, pgContainer *dockertest.Resource, pgUser, pgPassword, pgDb string) (*db.DbManager, error) {
	hostAndPort := pgContainer.GetHostPort("5432/tcp")
	dsn := fmt.Sprintf("postgres://%s:%s@%s/%s", pgUser, pgPassword, hostAndPort, pgDb)

	var dbManager *db.DbManager
	if err := pool.Retry(func() error {
		ctx, cancelFunc := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancelFunc()

		var err error
		dbManager, err = db.NewDbManager(&dsn)
		if err != nil {
			return err
		}

		return dbManager.Ping(ctx)
	}); err != nil {
		return nil, fmt.Errorf("failed to connect to DB: %w", err)
	}

	err := dbManager.Migrate()
	if err != nil {
		return nil, fmt.Errorf("failed to apply migrations: %w", err)
	}

	return dbManager, nil
}

type DockerTestSuite struct {
	Pool        *dockertest.Pool
	PGContainer *dockertest.Resource
	DBManager   *db.DbManager
}

func NewDockerTestSuite(t *testing.T) *DockerTestSuite {
	pool, err := CreateDockerPool()
	if err != nil {
		t.Fatalf("failed to connect to docker: %v", err)
	}

	pgUser := "admin"
	pgPassword := "123"
	pgDb := "sbtest"

	pgContainer, err := CreatePostgresContainer(t, pool, pgUser, pgPassword, pgDb)
	if err != nil {
		t.Fatalf("failed to create pg container: %v", err)
	}

	dbManager, err := CreateDbManager(pool, pgContainer, pgUser, pgPassword, pgDb)
	if err != nil {
		t.Fatalf("failed to create db manager: %v", err)
	}

	t.Cleanup(func() {
		dbManager.Close()
	})

	return &DockerTestSuite{
		Pool:        pool,
		PGContainer: pgContainer,
		DBManager:   dbManager,
	}
}

func CreateTestTheme(ctx context.Context, txm db.TxManager, nRounds int) (*qg.Theme, error) {
	return db.WithTxVQ(ctx, txm, func(ctx context.Context, q qg.Querier) (*qg.Theme, error) {
		theme, err := q.CreateTheme(ctx, "Test Theme", pgtype.Text{})
		if err != nil {
			return nil, err
		}

		for i := range nRounds {
			nRound := int32(i + 1)
			_, err = q.CreateRound(ctx, qg.CreateRoundParams{
				RoundN:  nRound,
				Title:   fmt.Sprintf("Round %d", nRound),
				ThemeID: theme.ID,
			})

			if err != nil {
				return nil, err
			}
		}

		return &theme, nil
	})
}

func CreateTestPlayers(ctx context.Context, txm db.TxManager, nPlayers int, tgIdOffset int) ([]qg.Player, error) {
	return db.WithTxVQ(ctx, txm, func(ctx context.Context, q qg.Querier) ([]qg.Player, error) {
		players := []qg.Player{}
		for i := range nPlayers {
			player, err := q.CreatePlayer(ctx, qg.CreatePlayerParams{
				TgID:       int64(tgIdOffset + i),
				TgUsername: fmt.Sprintf("player_%d", i),
			})
			if err != nil {
				return players, err
			}

			players = append(players, player)
		}

		return players, nil
	})
}

func GetGameStatus(ctx context.Context, txm db.TxManager, gameId int64) (qg.TestGetGameStatusByIdRow, error) {
	return db.WithTxVQ(ctx, txm, func(ctx context.Context, q qg.Querier) (qg.TestGetGameStatusByIdRow, error) {
		return q.TestGetGameStatusById(ctx, gameId)
	})
}

func CreateGameWithStage(ctx context.Context, txm db.TxManager, themeId int64, stage qg.GameStage) (*qg.Game, error) {
	return db.WithTxVQ(ctx, txm, func(ctx context.Context, q qg.Querier) (*qg.Game, error) {
		game, err := q.TestCreateGame(ctx, themeId)
		if err != nil {
			return nil, err
		}

		_, err = q.TestCreateGameStatus(ctx, qg.TestCreateGameStatusParams{
			GameID:  game.ID,
			Stage:   stage,
			ThemeID: game.ThemeID,
		})
		if err != nil {
			return nil, err
		}

		return &game, nil
	})
}

func ChangeGameStage(ctx context.Context, txm db.TxManager, gameId int64, stage qg.GameStage) (*qg.GameStatus, error) {
	return db.WithTxVQ(ctx, txm, func(ctx context.Context, q qg.Querier) (*qg.GameStatus, error) {
		status, err := q.TestUpdateGameStage(ctx, stage, gameId)
		if err != nil {
			return nil, err
		}

		return &status, nil
	})
}

func AddPlayerToGame(ctx context.Context, txm db.TxManager, gameId, playerId int64, mode qg.PlayerGameMode) (*qg.GamePlayer, error) {
	return db.WithTxVQ(ctx, txm, func(ctx context.Context, q qg.Querier) (*qg.GamePlayer, error) {
		gp, err := q.TestAddPlayerToGame(ctx, qg.TestAddPlayerToGameParams{
			GameID:   gameId,
			PlayerID: playerId,
			Mode:     mode,
		})

		if err != nil {
			return nil, err
		}

		return &gp, nil
	})
}
