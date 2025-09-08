package services

import (
	"context"
	"testing"

	"github.com/antonaby/shortsbattle/game-server/internal/db"
	"github.com/antonaby/shortsbattle/game-server/internal/db/qg"
	"github.com/antonaby/shortsbattle/game-server/internal/tests"
	"github.com/jackc/pgx/v5"
	"github.com/ory/dockertest/v3"
	"github.com/stretchr/testify/assert"
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

	t.Run("HandleLobby", func(t *testing.T) {
		err := db.WithTx(context.Background(), ts.dbManager, func(ctx context.Context, tx pgx.Tx) error {
			q := ts.dbManager.Querier(tx)
			_, err := q.CreatePlayer(ctx, qg.CreatePlayerParams{
				TgID: 1,
			})

			return err
		})

		if err != nil {
			t.Fatalf("failed to create user")
		}

		player, err := db.WithTxVQ(context.Background(), ts.dbManager, func(ctx context.Context, q qg.Querier) (qg.Player, error) {
			return q.GetPlayerByTgId(ctx, qg.GetPlayerByTgIdParams{
				TgID: 1,
			})
		})

		if err != nil {
			t.Fatalf("failed to create player")
		}

		assert.Equal(t, int64(1), player.TgID)
	})
}
