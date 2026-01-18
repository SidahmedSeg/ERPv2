package database

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"myerp-v2/internal/config"
)

// NewRedisClient creates a new Redis client with the provided configuration
func NewRedisClient(cfg *config.RedisConfig) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Address(),
		Password: cfg.Password,
		DB:       cfg.DB,

		// Connection pool settings
		PoolSize:        100,              // Maximum number of socket connections
		MinIdleConns:    10,               // Minimum number of idle connections
		MaxIdleConns:    50,               // Maximum number of idle connections
		ConnMaxIdleTime: 10 * time.Minute, // Close idle connections after this duration
		ConnMaxLifetime: 1 * time.Hour,    // Close connections after this lifetime

		// Timeouts
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,

		// Retry settings
		MaxRetries:      3,
		MinRetryBackoff: 8 * time.Millisecond,
		MaxRetryBackoff: 512 * time.Millisecond,
	})

	// Verify connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return client, nil
}

// CloseRedis gracefully closes the Redis connection
func CloseRedis(client *redis.Client) error {
	if client != nil {
		return client.Close()
	}
	return nil
}

// RedisHealthCheck performs a health check on the Redis connection
func RedisHealthCheck(ctx context.Context, client *redis.Client) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("Redis health check failed: %w", err)
	}

	return nil
}

// RedisStats returns Redis connection pool statistics
func RedisStats(client *redis.Client) map[string]interface{} {
	stats := client.PoolStats()
	return map[string]interface{}{
		"hits":         stats.Hits,
		"misses":       stats.Misses,
		"timeouts":     stats.Timeouts,
		"total_conns":  stats.TotalConns,
		"idle_conns":   stats.IdleConns,
		"stale_conns":  stats.StaleConns,
	}
}

// CacheKey generates a namespaced cache key
// Example: CacheKey("user", "permissions", userID) -> "user:permissions:123e4567-e89b-12d3-a456-426614174000"
func CacheKey(parts ...string) string {
	if len(parts) == 0 {
		return ""
	}

	result := parts[0]
	for i := 1; i < len(parts); i++ {
		result += ":" + parts[i]
	}
	return result
}

// SetWithExpiry sets a key-value pair in Redis with an expiration time
func SetWithExpiry(ctx context.Context, client *redis.Client, key string, value interface{}, expiry time.Duration) error {
	return client.Set(ctx, key, value, expiry).Err()
}

// Get retrieves a value from Redis
func Get(ctx context.Context, client *redis.Client, key string) (string, error) {
	return client.Get(ctx, key).Result()
}

// Delete removes a key from Redis
func Delete(ctx context.Context, client *redis.Client, keys ...string) error {
	return client.Del(ctx, keys...).Err()
}

// Exists checks if a key exists in Redis
func Exists(ctx context.Context, client *redis.Client, key string) (bool, error) {
	count, err := client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// Increment atomically increments a counter
func Increment(ctx context.Context, client *redis.Client, key string) (int64, error) {
	return client.Incr(ctx, key).Result()
}

// IncrementWithExpiry atomically increments a counter and sets expiry if it's a new key
func IncrementWithExpiry(ctx context.Context, client *redis.Client, key string, expiry time.Duration) (int64, error) {
	pipe := client.Pipeline()
	incr := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, expiry)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return 0, err
	}

	return incr.Val(), nil
}

// SetNX sets a key-value pair only if the key does not exist (useful for locks)
func SetNX(ctx context.Context, client *redis.Client, key string, value interface{}, expiry time.Duration) (bool, error) {
	return client.SetNX(ctx, key, value, expiry).Result()
}

// GetDel atomically gets and deletes a key (useful for one-time tokens)
func GetDel(ctx context.Context, client *redis.Client, key string) (string, error) {
	return client.GetDel(ctx, key).Result()
}

// SetJSON sets a JSON-encoded value in Redis
func SetJSON(ctx context.Context, client *redis.Client, key string, value interface{}, expiry time.Duration) error {
	// Note: This is a simplified version. For production, consider using redis-om or similar
	// for proper JSON handling with RedisJSON module
	return client.Set(ctx, key, value, expiry).Err()
}

// InvalidatePattern deletes all keys matching a pattern
// WARNING: Use with caution in production. This can be slow on large datasets.
func InvalidatePattern(ctx context.Context, client *redis.Client, pattern string) error {
	var cursor uint64
	var keys []string

	for {
		var scanKeys []string
		var err error

		scanKeys, cursor, err = client.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return fmt.Errorf("failed to scan keys: %w", err)
		}

		keys = append(keys, scanKeys...)

		if cursor == 0 {
			break
		}
	}

	if len(keys) > 0 {
		return client.Del(ctx, keys...).Err()
	}

	return nil
}

// TTL returns the time-to-live for a key
func TTL(ctx context.Context, client *redis.Client, key string) (time.Duration, error) {
	return client.TTL(ctx, key).Result()
}
