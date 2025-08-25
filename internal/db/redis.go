package db

import (
	"context"
	"errors"
	"os"
	"strings"
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

	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)

	_, err := client.Ping(ctx).Result()
	if err != nil {
		return nil, err
	}

	return client, nil
}

func EnsureStreamGroup(client *redis.Client, streamName, consumerGroupName string) error {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err := client.XGroupCreateMkStream(ctx, streamName, consumerGroupName, "0").Err()
	if err != nil && !strings.Contains(err.Error(), "BUSYGROUP") {
		return err
	}

	return nil
}
