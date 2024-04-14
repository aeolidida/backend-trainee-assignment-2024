package redis

import (
	"backend-trainee-assignment-2024/internal/config"
	"backend-trainee-assignment-2024/internal/errs"
	"context"
	"strconv"
	"time"

	redis "github.com/redis/go-redis/v9"
)

type Redis struct {
	client      *redis.Client
	ConnTimeout time.Duration
}

func NewRedis(cfg config.Redis) *Redis {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Address + ":" + strconv.Itoa(cfg.Port),
		Password: cfg.Password,
		DB:       0,
	})

	return &Redis{
		client:      client,
		ConnTimeout: cfg.ConnTimeout,
	}
}

func (c *Redis) Push(key, value string, ttl time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), c.ConnTimeout)
	defer cancel()

	return c.client.Set(ctx, key, value, ttl).Err()
}

func (c *Redis) Get(key string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.ConnTimeout)
	defer cancel()

	value, err := c.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", errs.ErrNotFound
	}
	return value, err
}
func (c *Redis) Remove(key string) error {
	ctx, cancel := context.WithTimeout(context.Background(), c.ConnTimeout)
	defer cancel()

	return c.client.Del(ctx, key).Err()
}

func (c *Redis) Close() error {
	return c.client.Close()
}
