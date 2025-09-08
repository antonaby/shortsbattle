package tests

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/antonaby/shortsbattle/game-server/internal/db"
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
		dbManager, err = db.NewDbManager(ctx, &dsn)
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
