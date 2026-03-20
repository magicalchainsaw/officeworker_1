package redis

import (
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type Blacklist struct {
	client *redis.Client
}

func NewBlacklist(client *redis.Client) *Blacklist {
	return &Blacklist{client: client}
}

func (b *Blacklist) Add(token string, expiration time.Duration) error {
	return b.client.Set(ctx, b.key(token), "1", expiration).Err()
}

func (b *Blacklist) Exists(token string) (bool, error) {
	count, err := b.client.Exists(ctx, b.key(token)).Result()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (b *Blacklist) Remove(token string) error {
	return b.client.Del(ctx, b.key(token)).Err()
}

func (b *Blacklist) key(token string) string {
	return fmt.Sprintf("blacklist:%s", token)
}
