package database

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"

	"github.com/reazonvan/LootBay/internal/config"
)

// RedisDB обертка для работы с Redis
type RedisDB struct {
	Client *redis.Client
}

// NewRedisDB создает новый экземпляр RedisDB
func NewRedisDB(cfg *config.RedisConfig) (*RedisDB, error) {
	client, err := NewRedisConnection(cfg)
	if err != nil {
		return nil, err
	}
	return &RedisDB{Client: client}, nil
}

// Close закрывает подключение к Redis
func (r *RedisDB) Close() error {
	return CloseRedis(r.Client)
}

// NewRedisConnection создает подключение к Redis
func NewRedisConnection(cfg *config.RedisConfig) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Password:     cfg.Password,
		DB:           cfg.DB,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConns,
	})

	// Проверка подключения
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return client, nil
}

// CloseRedis закрывает подключение к Redis
func CloseRedis(client *redis.Client) error {
	return client.Close()
}
