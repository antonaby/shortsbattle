package services

import (
	"context"
	"testing"

	"github.com/antonaby/shortsbattle/game-server/internal/db"
	"github.com/antonaby/shortsbattle/game-server/internal/db/qg"
	"github.com/antonaby/shortsbattle/game-server/internal/tests"
	"github.com/jackc/pgx/v5"
)

func TestAdvanceGame(t *testing.T) {
	pool, err := tests.CreateDockerPool()
	if err != nil {
		t.Fatalf("could not connect to docker: %v", err)
	}

	pgUser := "admin"
	pgPassword := "123"
	pgDb := "sbtest"

	pgContainer, err := tests.CreatePostgresContainer(t, pool, pgUser, pgPassword, pgDb)
	if err != nil {
		t.Fatalf("could not create Postgres container: %v", err)
	}

	dbManager, err := tests.CreateDbManager(pool, pgContainer, pgUser, pgPassword, pgDb)
	if err != nil {
		t.Fatalf("failed to create db manager: %v", err)
	}

	defer dbManager.Close()

	err = db.WithTx(context.Background(), dbManager, func(ctx context.Context, tx pgx.Tx) error {
		q := dbManager.Querier(tx)
		_, err := q.CreatePlayer(ctx, qg.CreatePlayerParams{
			TgID: 1,
		})

		return err
	})

	if err != nil {
		t.Fatalf("failed to create user")
	}

	player, err := db.WithTxVQ(context.Background(), dbManager, func(ctx context.Context, q qg.Querier) (qg.Player, error) {
		return q.GetPlayerByTgId(ctx, qg.GetPlayerByTgIdParams{
			TgID: 1,
		})
	})

	if err != nil {
		t.Fatalf("failed to create fetch")
	}

	if player.TgID != 1 {
		t.Errorf("incorrect tgId")
	}
}
