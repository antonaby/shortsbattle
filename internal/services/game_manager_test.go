package services

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/antonaby/shortsbattle/game-server/internal/db"
	"github.com/ory/dockertest/v3"
)

func TestAdvanceGame(t *testing.T) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		t.Fatalf("could not connect to docker: %v", err)
	}

	err = pool.Client.Ping()
	if err != nil {
		t.Fatalf("could not connect to Docker: %v", err)
	}

	resource, err := pool.Run("postgres", "17", []string{"POSTGRES_USER=admin", "POSTGRES_PASSWORD=123", "POSTGRES_DB=sbtest"})
	if err != nil {
		t.Fatalf("could not create Postgres container: %v", err)
	}

	t.Cleanup(func() {
		if err := pool.Purge(resource); err != nil {
			log.Printf("failed to purge docker resource: %v", err)
		}
	})

	hostAndPort := resource.GetHostPort("5432/tcp")
	dsn := fmt.Sprintf("postgres://%s:%s@%s/%s", "admin", "123", hostAndPort, "sbtest")

	var dbManager *db.DbManager
	if err := pool.Retry(func() error {
		ctx, cancelFunc := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancelFunc()

		var err error
		dbManager, err = db.NewDbManager(ctx, &dsn)
		if err != nil {
			return err
		}

		return dbManager.Ping(ctx)
	}); err != nil {
		log.Fatalf("Could not connect to database: %s", err)
	}

	defer dbManager.Close()
}
