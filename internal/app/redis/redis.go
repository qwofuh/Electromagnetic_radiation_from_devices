package redis

import (
	"context"
	"fmt"
	cfg "lab1/internal/app/config"
	"strconv"

	"github.com/go-redis/redis/v8"
)

const servicePrefix = "user_auth_service."

type Client struct {
	cfg    cfg.RedisConfig
	client *redis.Client
}

func New(ctx context.Context, cfgIn cfg.RedisConfig) (*Client, error) {
	fmt.Printf("Redis connection: %s:%d, password: '%s'\n", cfgIn.Host, cfgIn.Port, cfgIn.Password)
	client := &Client{}

	client.cfg = cfgIn

	redisClient := redis.NewClient(&redis.Options{
		Password:    cfgIn.Password,
		Username:    cfgIn.User,
		Addr:        cfgIn.Host + ":" + strconv.Itoa(cfgIn.Port),
		DB:          0,
		DialTimeout: cfgIn.DialTimeout,
		ReadTimeout: cfgIn.ReadTimeout,
	})

	client.client = redisClient

	if _, err := redisClient.Ping(ctx).Result(); err != nil {
		return nil, fmt.Errorf("cant ping redis: %w", err)
	}

	return client, nil
}

func (c *Client) Close() error {
	return c.client.Close()
}
