package redis

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()

type Client struct {
	client *redis.Client
}

type Config struct {
	Host     string
	Port     int
	Password string
	DB       int
}

func New(cfg *Config) (*Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to ping redis: %w", err)
	}

	return &Client{client: rdb}, nil
}

func (c *Client) GetClient() *redis.Client {
	return c.client
}

func (c *Client) Close() error {
	return c.client.Close()
}
