package database

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/digi604/swarmmarket/backend/internal/config"
)

// RedisDB wraps a Redis client.
type RedisDB struct {
	Client *redis.Client
}

// NewRedisDB creates a new Redis client.
func NewRedisDB(ctx context.Context, cfg config.RedisConfig) (*RedisDB, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Address(),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	// Verify connection
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to ping redis: %w", err)
	}

	return &RedisDB{Client: client}, nil
}

// Close closes the Redis client.
func (r *RedisDB) Close() error {
	if r.Client != nil {
		return r.Client.Close()
	}
	return nil
}

// Health checks if Redis is reachable.
func (r *RedisDB) Health(ctx context.Context) error {
	return r.Client.Ping(ctx).Err()
}

// Publish publishes a message to a Redis channel.
func (r *RedisDB) Publish(ctx context.Context, channel string, message interface{}) error {
	return r.Client.Publish(ctx, channel, message).Err()
}

// Subscribe subscribes to Redis channels.
func (r *RedisDB) Subscribe(ctx context.Context, channels ...string) *redis.PubSub {
	return r.Client.Subscribe(ctx, channels...)
}
