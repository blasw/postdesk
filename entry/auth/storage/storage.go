package storage

import (
	"context"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
)

type RedisStorage struct {
	client *redis.Client
}

func NewRedisStorage(addr string) (*RedisStorage, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "",
		DB:       0,
	})

	ctx := context.Background()

	var err error

	// Trying to connect to redis 10 times with 1 second timeout
	for i := 0; i < 10; i++ {
		_, err = rdb.Ping(ctx).Result()
		if err == nil {
			logrus.Debug("Storage connected")
			break
		}
		logrus.Debug("Failed to connect to storage. Try: ", i+1)
		time.Sleep(time.Second)
	}

	if err != nil {
		logrus.Debug("Unable to connect to storage")
		return nil, err
	}

	return &RedisStorage{
		client: rdb,
	}, nil
}

// StoreToken stores provided refresh token in a redis cache. Key is userID and value is token.
func (s *RedisStorage) SetToken(userID uint64, token string) error {
	return s.client.Set(context.Background(), strconv.Itoa(int(userID)), token, 30*24*time.Hour).Err()
}

// GetToken returns refresh token for provided userID.
func (s *RedisStorage) GetToken(userID uint64) (string, error) {
	return s.client.Get(context.Background(), strconv.Itoa(int(userID))).Result()
}
