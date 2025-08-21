package db

import (
	"context"
	"errors"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

var ErrorRedisHostNotDefined = errors.New("REDIS_HOST not defined")

func NewRedisClient() (*redis.Client, error) {
	redisHost := os.Getenv("REDIS_HOST")
	if len(redisHost) == 0 {
		return nil, ErrorRedisHostNotDefined
	}

	client := redis.NewClient(&redis.Options{
		Addr:     redisHost,
		Password: "",
		DB:       0,
	})

	ctx, _ := context.WithTimeout(context.Background(), 10 * time.Second)

	_, err := client.Ping(ctx).Result()
	if err != nil {
		return nil, err
	}

	return client, nil
}