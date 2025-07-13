package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/reazonvan/LootBay/internal/config"
	"github.com/reazonvan/LootBay/pkg/logger"
	"github.com/redis/go-redis/v9"
)

// Cache интерфейс для кэширования
type Cache interface {
	Get(ctx context.Context, key string, dest interface{}) error
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
	Incr(ctx context.Context, key string) (int64, error)
	IncrBy(ctx context.Context, key string, value int64) (int64, error)
	Expire(ctx context.Context, key string, expiration time.Duration) error
	TTL(ctx context.Context, key string) (time.Duration, error)
	Flush(ctx context.Context) error
	GetKeys(ctx context.Context, pattern string) ([]string, error)
}

// DistributedLock интерфейс для distributed locking
type DistributedLock interface {
	Lock(ctx context.Context, key string, expiration time.Duration) (bool, error)
	Unlock(ctx context.Context, key string) error
	TryLock(ctx context.Context, key string, expiration time.Duration) (bool, error)
}

// RedisCache реализация кэша с Redis
type RedisCache struct {
	client *redis.Client
	logger *logger.StructuredLogger
}

// NewRedisCache создает новый Redis кэш
func NewRedisCache(cfg *config.Config, log *logger.StructuredLogger) (*RedisCache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port),
		Password:     cfg.Redis.Password,
		DB:           cfg.Redis.DB,
		PoolSize:     cfg.Redis.PoolSize,
		MinIdleConns: cfg.Redis.MinIdleConns,
	})

	// Проверяем подключение
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	log.Info("Redis cache initialized", map[string]interface{}{
		"host": cfg.Redis.Host,
		"port": cfg.Redis.Port,
		"db":   cfg.Redis.DB,
	})

	return &RedisCache{
		client: client,
		logger: log,
	}, nil
}

// Get получает значение из кэша
func (rc *RedisCache) Get(ctx context.Context, key string, dest interface{}) error {
	start := time.Now()

	data, err := rc.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			rc.logger.LogCacheOperation(ctx, "get", key, time.Since(start), false, nil)
			return fmt.Errorf("key not found: %s", key)
		}
		rc.logger.LogCacheOperation(ctx, "get", key, time.Since(start), false, err)
		return fmt.Errorf("failed to get key %s: %w", key, err)
	}

	if err := json.Unmarshal(data, dest); err != nil {
		rc.logger.LogCacheOperation(ctx, "get", key, time.Since(start), false, err)
		return fmt.Errorf("failed to unmarshal data for key %s: %w", key, err)
	}

	rc.logger.LogCacheOperation(ctx, "get", key, time.Since(start), true, nil)
	return nil
}

// Set устанавливает значение в кэш
func (rc *RedisCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	start := time.Now()

	data, err := json.Marshal(value)
	if err != nil {
		rc.logger.LogCacheOperation(ctx, "set", key, time.Since(start), false, err)
		return fmt.Errorf("failed to marshal value for key %s: %w", key, err)
	}

	err = rc.client.Set(ctx, key, data, expiration).Err()
	if err != nil {
		rc.logger.LogCacheOperation(ctx, "set", key, time.Since(start), false, err)
		return fmt.Errorf("failed to set key %s: %w", key, err)
	}

	rc.logger.LogCacheOperation(ctx, "set", key, time.Since(start), true, nil)
	return nil
}

// Delete удаляет значение из кэша
func (rc *RedisCache) Delete(ctx context.Context, key string) error {
	start := time.Now()

	err := rc.client.Del(ctx, key).Err()
	if err != nil {
		rc.logger.LogCacheOperation(ctx, "delete", key, time.Since(start), false, err)
		return fmt.Errorf("failed to delete key %s: %w", key, err)
	}

	rc.logger.LogCacheOperation(ctx, "delete", key, time.Since(start), true, nil)
	return nil
}

// Exists проверяет существование ключа
func (rc *RedisCache) Exists(ctx context.Context, key string) (bool, error) {
	start := time.Now()

	exists, err := rc.client.Exists(ctx, key).Result()
	if err != nil {
		rc.logger.LogCacheOperation(ctx, "exists", key, time.Since(start), false, err)
		return false, fmt.Errorf("failed to check existence of key %s: %w", key, err)
	}

	rc.logger.LogCacheOperation(ctx, "exists", key, time.Since(start), true, nil)
	return exists > 0, nil
}

// Incr увеличивает значение на 1
func (rc *RedisCache) Incr(ctx context.Context, key string) (int64, error) {
	start := time.Now()

	value, err := rc.client.Incr(ctx, key).Result()
	if err != nil {
		rc.logger.LogCacheOperation(ctx, "incr", key, time.Since(start), false, err)
		return 0, fmt.Errorf("failed to increment key %s: %w", key, err)
	}

	rc.logger.LogCacheOperation(ctx, "incr", key, time.Since(start), true, nil)
	return value, nil
}

// IncrBy увеличивает значение на указанное количество
func (rc *RedisCache) IncrBy(ctx context.Context, key string, value int64) (int64, error) {
	start := time.Now()

	result, err := rc.client.IncrBy(ctx, key, value).Result()
	if err != nil {
		rc.logger.LogCacheOperation(ctx, "incrby", key, time.Since(start), false, err)
		return 0, fmt.Errorf("failed to increment key %s by %d: %w", key, value, err)
	}

	rc.logger.LogCacheOperation(ctx, "incrby", key, time.Since(start), true, nil)
	return result, nil
}

// Expire устанавливает время жизни ключа
func (rc *RedisCache) Expire(ctx context.Context, key string, expiration time.Duration) error {
	start := time.Now()

	err := rc.client.Expire(ctx, key, expiration).Err()
	if err != nil {
		rc.logger.LogCacheOperation(ctx, "expire", key, time.Since(start), false, err)
		return fmt.Errorf("failed to set expiration for key %s: %w", key, err)
	}

	rc.logger.LogCacheOperation(ctx, "expire", key, time.Since(start), true, nil)
	return nil
}

// TTL получает оставшееся время жизни ключа
func (rc *RedisCache) TTL(ctx context.Context, key string) (time.Duration, error) {
	start := time.Now()

	ttl, err := rc.client.TTL(ctx, key).Result()
	if err != nil {
		rc.logger.LogCacheOperation(ctx, "ttl", key, time.Since(start), false, err)
		return 0, fmt.Errorf("failed to get TTL for key %s: %w", key, err)
	}

	rc.logger.LogCacheOperation(ctx, "ttl", key, time.Since(start), true, nil)
	return ttl, nil
}

// Flush очищает весь кэш
func (rc *RedisCache) Flush(ctx context.Context) error {
	start := time.Now()

	err := rc.client.FlushDB(ctx).Err()
	if err != nil {
		rc.logger.LogCacheOperation(ctx, "flush", "all", time.Since(start), false, err)
		return fmt.Errorf("failed to flush cache: %w", err)
	}

	rc.logger.LogCacheOperation(ctx, "flush", "all", time.Since(start), true, nil)
	return nil
}

// GetKeys получает ключи по паттерну
func (rc *RedisCache) GetKeys(ctx context.Context, pattern string) ([]string, error) {
	start := time.Now()

	keys, err := rc.client.Keys(ctx, pattern).Result()
	if err != nil {
		rc.logger.LogCacheOperation(ctx, "keys", pattern, time.Since(start), false, err)
		return nil, fmt.Errorf("failed to get keys for pattern %s: %w", pattern, err)
	}

	rc.logger.LogCacheOperation(ctx, "keys", pattern, time.Since(start), true, nil)
	return keys, nil
}

// Close закрывает соединение с Redis
func (rc *RedisCache) Close() error {
	return rc.client.Close()
}

// RedisDistributedLock реализация distributed locking с Redis
type RedisDistributedLock struct {
	client *redis.Client
	logger *logger.StructuredLogger
}

// NewRedisDistributedLock создает новый distributed lock
func NewRedisDistributedLock(cfg *config.Config, log *logger.StructuredLogger) (*RedisDistributedLock, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port),
		Password:     cfg.Redis.Password,
		DB:           cfg.Redis.DB,
		PoolSize:     cfg.Redis.PoolSize,
		MinIdleConns: cfg.Redis.MinIdleConns,
	})

	// Проверяем подключение
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisDistributedLock{
		client: client,
		logger: log,
	}, nil
}

// Lock блокирует ресурс (блокирующий вызов)
func (rdl *RedisDistributedLock) Lock(ctx context.Context, key string, expiration time.Duration) (bool, error) {
	lockKey := fmt.Sprintf("lock:%s", key)

	// Пытаемся получить блокировку
	for {
		select {
		case <-ctx.Done():
			return false, ctx.Err()
		default:
			// Используем SET с NX (только если ключ не существует) и EX (expiration)
			result, err := rdl.client.SetNX(ctx, lockKey, "locked", expiration).Result()
			if err != nil {
				return false, fmt.Errorf("failed to acquire lock for key %s: %w", key, err)
			}

			if result {
				rdl.logger.Info("Lock acquired", map[string]interface{}{
					"key":        key,
					"expiration": expiration.String(),
				})
				return true, nil
			}

			// Ждем немного перед повторной попыткой
			time.Sleep(100 * time.Millisecond)
		}
	}
}

// TryLock пытается получить блокировку (неблокирующий вызов)
func (rdl *RedisDistributedLock) TryLock(ctx context.Context, key string, expiration time.Duration) (bool, error) {
	lockKey := fmt.Sprintf("lock:%s", key)

	result, err := rdl.client.SetNX(ctx, lockKey, "locked", expiration).Result()
	if err != nil {
		return false, fmt.Errorf("failed to try acquire lock for key %s: %w", key, err)
	}

	if result {
		rdl.logger.Info("Lock acquired (try)", map[string]interface{}{
			"key":        key,
			"expiration": expiration.String(),
		})
	}

	return result, nil
}

// Unlock освобождает блокировку
func (rdl *RedisDistributedLock) Unlock(ctx context.Context, key string) error {
	lockKey := fmt.Sprintf("lock:%s", key)

	// Используем Lua скрипт для атомарного удаления
	script := `
		if redis.call("get", KEYS[1]) == ARGV[1] then
			return redis.call("del", KEYS[1])
		else
			return 0
		end
	`

	result, err := rdl.client.Eval(ctx, script, []string{lockKey}, "locked").Result()
	if err != nil {
		return fmt.Errorf("failed to release lock for key %s: %w", key, err)
	}

	if result.(int64) == 1 {
		rdl.logger.Info("Lock released", map[string]interface{}{
			"key": key,
		})
	} else {
		rdl.logger.Warn("Lock not found or already released", map[string]interface{}{
			"key": key,
		})
	}

	return nil
}

// Close закрывает соединение с Redis
func (rdl *RedisDistributedLock) Close() error {
	return rdl.client.Close()
}

// CacheManager менеджер кэширования с дополнительными возможностями
type CacheManager struct {
	cache  Cache
	lock   DistributedLock
	logger *logger.StructuredLogger
}

// NewCacheManager создает новый менеджер кэширования
func NewCacheManager(cfg *config.Config, log *logger.StructuredLogger) (*CacheManager, error) {
	cache, err := NewRedisCache(cfg, log)
	if err != nil {
		return nil, err
	}

	lock, err := NewRedisDistributedLock(cfg, log)
	if err != nil {
		return nil, err
	}

	return &CacheManager{
		cache:  cache,
		lock:   lock,
		logger: log,
	}, nil
}

// GetWithLock получает значение с блокировкой
func (cm *CacheManager) GetWithLock(ctx context.Context, key string, dest interface{}, lockTimeout time.Duration) error {
	// Пытаемся получить блокировку
	locked, err := cm.lock.TryLock(ctx, key, lockTimeout)
	if err != nil {
		return fmt.Errorf("failed to acquire lock: %w", err)
	}

	if !locked {
		return fmt.Errorf("failed to acquire lock for key %s", key)
	}

	defer func() {
		_ = cm.lock.Unlock(ctx, key)
	}()

	// Получаем значение из кэша
	return cm.cache.Get(ctx, key, dest)
}

// SetWithInvalidation устанавливает значение и инвалидирует связанные ключи
func (cm *CacheManager) SetWithInvalidation(ctx context.Context, key string, value interface{}, expiration time.Duration, invalidatePatterns []string) error {
	// Устанавливаем значение
	if err := cm.cache.Set(ctx, key, value, expiration); err != nil {
		return err
	}

	// Инвалидируем связанные ключи
	for _, pattern := range invalidatePatterns {
		keys, err := cm.cache.GetKeys(ctx, pattern)
		if err != nil {
			cm.logger.Warn("Failed to get keys for invalidation", map[string]interface{}{
				"pattern": pattern,
				"error":   err.Error(),
			})
			continue
		}

		for _, invalidateKey := range keys {
			if invalidateKey != key { // Не инвалидируем текущий ключ
				if err := cm.cache.Delete(ctx, invalidateKey); err != nil {
					cm.logger.Warn("Failed to invalidate key", map[string]interface{}{
						"key":   invalidateKey,
						"error": err.Error(),
					})
				}
			}
		}
	}

	return nil
}

// GetCache возвращает интерфейс кэша
func (cm *CacheManager) GetCache() Cache {
	return cm.cache
}

// GetLock возвращает интерфейс блокировки
func (cm *CacheManager) GetLock() DistributedLock {
	return cm.lock
}

// Close закрывает все соединения
func (cm *CacheManager) Close() error {
	var errors []error

	if err := cm.cache.(*RedisCache).Close(); err != nil {
		errors = append(errors, fmt.Errorf("failed to close cache: %w", err))
	}

	if err := cm.lock.(*RedisDistributedLock).Close(); err != nil {
		errors = append(errors, fmt.Errorf("failed to close lock: %w", err))
	}

	if len(errors) > 0 {
		return fmt.Errorf("errors closing cache manager: %v", errors)
	}

	return nil
}
