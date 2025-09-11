package db

import (
	"context"
	"errors"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

var ErrorRedisHostNotDefined = errors.New("REDIS_HOST not defined")

func GetRedisDSN() (string, error) {
	redisHost := os.Getenv("REDIS_HOST")
	if len(redisHost) == 0 {
		return "", ErrorRedisHostNotDefined
	}

	return redisHost, nil
}

func NewRedisClient(redisHost string) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     redisHost,
		Password: "",
		DB:       0,
	})

	ctx, cancelFunc := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelFunc()

	_, err := client.Ping(ctx).Result()
	if err != nil {
		return nil, err
	}

	return client, nil
}
