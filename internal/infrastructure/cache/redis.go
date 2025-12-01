package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// RedisClient wraps the Redis client with additional functionality
type RedisClient struct {
	client *redis.Client
	logger *zap.Logger
	prefix string
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Host            string
	Port            string
	Password        string
	DB              int
	MaxRetries      int
	PoolSize        int
	MinIdleConns    int
	DialTimeout     time.Duration
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	PoolTimeout     time.Duration
	ConnMaxIdleTime time.Duration
	KeyPrefix       string
}

// NewRedisClient creates a new Redis client
func NewRedisClient(config RedisConfig, logger *zap.Logger) (*RedisClient, error) {
	addr := fmt.Sprintf("%s:%s", config.Host, config.Port)

	client := redis.NewClient(&redis.Options{
		Addr:            addr,
		Password:        config.Password,
		DB:              config.DB,
		MaxRetries:      config.MaxRetries,
		PoolSize:        config.PoolSize,
		MinIdleConns:    config.MinIdleConns,
		DialTimeout:     config.DialTimeout,
		ReadTimeout:     config.ReadTimeout,
		WriteTimeout:    config.WriteTimeout,
		PoolTimeout:     config.PoolTimeout,
		ConnMaxIdleTime: config.ConnMaxIdleTime,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	logger.Info("connected to Redis successfully",
		zap.String("address", addr),
		zap.Int("db", config.DB),
	)

	return &RedisClient{
		client: client,
		logger: logger,
		prefix: config.KeyPrefix,
	}, nil
}

// makeKey creates a prefixed key
func (r *RedisClient) makeKey(key string) string {
	if r.prefix == "" {
		return key
	}
	return fmt.Sprintf("%s:%s", r.prefix, key)
}

// Get retrieves a value from Redis
func (r *RedisClient) Get(ctx context.Context, key string) (string, error) {
	fullKey := r.makeKey(key)
	val, err := r.client.Get(ctx, fullKey).Result()
	if err == redis.Nil {
		return "", ErrCacheMiss
	}
	if err != nil {
		r.logger.Error("failed to get key from Redis",
			zap.String("key", fullKey),
			zap.Error(err),
		)
		return "", err
	}
	return val, nil
}

// Set stores a value in Redis with TTL
func (r *RedisClient) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	fullKey := r.makeKey(key)
	err := r.client.Set(ctx, fullKey, value, ttl).Err()
	if err != nil {
		r.logger.Error("failed to set key in Redis",
			zap.String("key", fullKey),
			zap.Error(err),
		)
		return err
	}
	return nil
}

// GetJSON retrieves and unmarshals a JSON value from Redis
func (r *RedisClient) GetJSON(ctx context.Context, key string, dest interface{}) error {
	val, err := r.Get(ctx, key)
	if err != nil {
		return err
	}

	if err := json.Unmarshal([]byte(val), dest); err != nil {
		r.logger.Error("failed to unmarshal JSON",
			zap.String("key", key),
			zap.Error(err),
		)
		return err
	}

	return nil
}

// SetJSON marshals and stores a JSON value in Redis
func (r *RedisClient) SetJSON(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		r.logger.Error("failed to marshal JSON",
			zap.String("key", key),
			zap.Error(err),
		)
		return err
	}

	return r.Set(ctx, key, data, ttl)
}

// Delete removes a key from Redis
func (r *RedisClient) Delete(ctx context.Context, keys ...string) error {
	fullKeys := make([]string, len(keys))
	for i, key := range keys {
		fullKeys[i] = r.makeKey(key)
	}

	err := r.client.Del(ctx, fullKeys...).Err()
	if err != nil {
		r.logger.Error("failed to delete keys from Redis",
			zap.Strings("keys", fullKeys),
			zap.Error(err),
		)
		return err
	}
	return nil
}

// DeletePattern deletes all keys matching a pattern
func (r *RedisClient) DeletePattern(ctx context.Context, pattern string) error {
	fullPattern := r.makeKey(pattern)

	iter := r.client.Scan(ctx, 0, fullPattern, 0).Iterator()
	var keys []string

	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}

	if err := iter.Err(); err != nil {
		r.logger.Error("failed to scan keys",
			zap.String("pattern", fullPattern),
			zap.Error(err),
		)
		return err
	}

	if len(keys) > 0 {
		err := r.client.Del(ctx, keys...).Err()
		if err != nil {
			r.logger.Error("failed to delete keys",
				zap.Int("count", len(keys)),
				zap.Error(err),
			)
			return err
		}

		r.logger.Info("deleted keys by pattern",
			zap.String("pattern", fullPattern),
			zap.Int("count", len(keys)),
		)
	}

	return nil
}

// Exists checks if a key exists in Redis
func (r *RedisClient) Exists(ctx context.Context, keys ...string) (int64, error) {
	fullKeys := make([]string, len(keys))
	for i, key := range keys {
		fullKeys[i] = r.makeKey(key)
	}

	count, err := r.client.Exists(ctx, fullKeys...).Result()
	if err != nil {
		r.logger.Error("failed to check key existence",
			zap.Strings("keys", fullKeys),
			zap.Error(err),
		)
		return 0, err
	}
	return count, nil
}

// Expire sets a TTL on an existing key
func (r *RedisClient) Expire(ctx context.Context, key string, ttl time.Duration) error {
	fullKey := r.makeKey(key)
	err := r.client.Expire(ctx, fullKey, ttl).Err()
	if err != nil {
		r.logger.Error("failed to set expiration",
			zap.String("key", fullKey),
			zap.Duration("ttl", ttl),
			zap.Error(err),
		)
		return err
	}
	return nil
}

// Increment increments a counter
func (r *RedisClient) Increment(ctx context.Context, key string) (int64, error) {
	fullKey := r.makeKey(key)
	val, err := r.client.Incr(ctx, fullKey).Result()
	if err != nil {
		r.logger.Error("failed to increment key",
			zap.String("key", fullKey),
			zap.Error(err),
		)
		return 0, err
	}
	return val, nil
}

// IncrementBy increments a counter by a specific amount
func (r *RedisClient) IncrementBy(ctx context.Context, key string, value int64) (int64, error) {
	fullKey := r.makeKey(key)
	val, err := r.client.IncrBy(ctx, fullKey, value).Result()
	if err != nil {
		r.logger.Error("failed to increment key by value",
			zap.String("key", fullKey),
			zap.Int64("value", value),
			zap.Error(err),
		)
		return 0, err
	}
	return val, nil
}

// Decrement decrements a counter
func (r *RedisClient) Decrement(ctx context.Context, key string) (int64, error) {
	fullKey := r.makeKey(key)
	val, err := r.client.Decr(ctx, fullKey).Result()
	if err != nil {
		r.logger.Error("failed to decrement key",
			zap.String("key", fullKey),
			zap.Error(err),
		)
		return 0, err
	}
	return val, nil
}

// SetNX sets a key only if it doesn't exist (SET if Not eXists)
func (r *RedisClient) SetNX(ctx context.Context, key string, value interface{}, ttl time.Duration) (bool, error) {
	fullKey := r.makeKey(key)
	ok, err := r.client.SetNX(ctx, fullKey, value, ttl).Result()
	if err != nil {
		r.logger.Error("failed to setnx",
			zap.String("key", fullKey),
			zap.Error(err),
		)
		return false, err
	}
	return ok, nil
}

// GetTTL returns the remaining TTL of a key
func (r *RedisClient) GetTTL(ctx context.Context, key string) (time.Duration, error) {
	fullKey := r.makeKey(key)
	ttl, err := r.client.TTL(ctx, fullKey).Result()
	if err != nil {
		r.logger.Error("failed to get TTL",
			zap.String("key", fullKey),
			zap.Error(err),
		)
		return 0, err
	}
	return ttl, nil
}

// HSet sets a field in a hash
func (r *RedisClient) HSet(ctx context.Context, key string, field string, value interface{}) error {
	fullKey := r.makeKey(key)
	err := r.client.HSet(ctx, fullKey, field, value).Err()
	if err != nil {
		r.logger.Error("failed to hset",
			zap.String("key", fullKey),
			zap.String("field", field),
			zap.Error(err),
		)
		return err
	}
	return nil
}

// HGet gets a field from a hash
func (r *RedisClient) HGet(ctx context.Context, key string, field string) (string, error) {
	fullKey := r.makeKey(key)
	val, err := r.client.HGet(ctx, fullKey, field).Result()
	if err == redis.Nil {
		return "", ErrCacheMiss
	}
	if err != nil {
		r.logger.Error("failed to hget",
			zap.String("key", fullKey),
			zap.String("field", field),
			zap.Error(err),
		)
		return "", err
	}
	return val, nil
}

// HGetAll gets all fields and values from a hash
func (r *RedisClient) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	fullKey := r.makeKey(key)
	val, err := r.client.HGetAll(ctx, fullKey).Result()
	if err != nil {
		r.logger.Error("failed to hgetall",
			zap.String("key", fullKey),
			zap.Error(err),
		)
		return nil, err
	}
	return val, nil
}

// HDel deletes fields from a hash
func (r *RedisClient) HDel(ctx context.Context, key string, fields ...string) error {
	fullKey := r.makeKey(key)
	err := r.client.HDel(ctx, fullKey, fields...).Err()
	if err != nil {
		r.logger.Error("failed to hdel",
			zap.String("key", fullKey),
			zap.Strings("fields", fields),
			zap.Error(err),
		)
		return err
	}
	return nil
}

// LPush pushes values to the head of a list
func (r *RedisClient) LPush(ctx context.Context, key string, values ...interface{}) error {
	fullKey := r.makeKey(key)
	err := r.client.LPush(ctx, fullKey, values...).Err()
	if err != nil {
		r.logger.Error("failed to lpush",
			zap.String("key", fullKey),
			zap.Error(err),
		)
		return err
	}
	return nil
}

// RPush pushes values to the tail of a list
func (r *RedisClient) RPush(ctx context.Context, key string, values ...interface{}) error {
	fullKey := r.makeKey(key)
	err := r.client.RPush(ctx, fullKey, values...).Err()
	if err != nil {
		r.logger.Error("failed to rpush",
			zap.String("key", fullKey),
			zap.Error(err),
		)
		return err
	}
	return nil
}

// LPop pops a value from the head of a list
func (r *RedisClient) LPop(ctx context.Context, key string) (string, error) {
	fullKey := r.makeKey(key)
	val, err := r.client.LPop(ctx, fullKey).Result()
	if err == redis.Nil {
		return "", ErrCacheMiss
	}
	if err != nil {
		r.logger.Error("failed to lpop",
			zap.String("key", fullKey),
			zap.Error(err),
		)
		return "", err
	}
	return val, nil
}

// RPop pops a value from the tail of a list
func (r *RedisClient) RPop(ctx context.Context, key string) (string, error) {
	fullKey := r.makeKey(key)
	val, err := r.client.RPop(ctx, fullKey).Result()
	if err == redis.Nil {
		return "", ErrCacheMiss
	}
	if err != nil {
		r.logger.Error("failed to rpop",
			zap.String("key", fullKey),
			zap.Error(err),
		)
		return "", err
	}
	return val, nil
}

// LRange returns a range of elements from a list
func (r *RedisClient) LRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	fullKey := r.makeKey(key)
	vals, err := r.client.LRange(ctx, fullKey, start, stop).Result()
	if err != nil {
		r.logger.Error("failed to lrange",
			zap.String("key", fullKey),
			zap.Error(err),
		)
		return nil, err
	}
	return vals, nil
}

// SAdd adds members to a set
func (r *RedisClient) SAdd(ctx context.Context, key string, members ...interface{}) error {
	fullKey := r.makeKey(key)
	err := r.client.SAdd(ctx, fullKey, members...).Err()
	if err != nil {
		r.logger.Error("failed to sadd",
			zap.String("key", fullKey),
			zap.Error(err),
		)
		return err
	}
	return nil
}

// SMembers returns all members of a set
func (r *RedisClient) SMembers(ctx context.Context, key string) ([]string, error) {
	fullKey := r.makeKey(key)
	members, err := r.client.SMembers(ctx, fullKey).Result()
	if err != nil {
		r.logger.Error("failed to smembers",
			zap.String("key", fullKey),
			zap.Error(err),
		)
		return nil, err
	}
	return members, nil
}

// SIsMember checks if a value is a member of a set
func (r *RedisClient) SIsMember(ctx context.Context, key string, member interface{}) (bool, error) {
	fullKey := r.makeKey(key)
	ok, err := r.client.SIsMember(ctx, fullKey, member).Result()
	if err != nil {
		r.logger.Error("failed to sismember",
			zap.String("key", fullKey),
			zap.Error(err),
		)
		return false, err
	}
	return ok, nil
}

// SRem removes members from a set
func (r *RedisClient) SRem(ctx context.Context, key string, members ...interface{}) error {
	fullKey := r.makeKey(key)
	err := r.client.SRem(ctx, fullKey, members...).Err()
	if err != nil {
		r.logger.Error("failed to srem",
			zap.String("key", fullKey),
			zap.Error(err),
		)
		return err
	}
	return nil
}

// ZAdd adds members with scores to a sorted set
func (r *RedisClient) ZAdd(ctx context.Context, key string, members ...redis.Z) error {
	fullKey := r.makeKey(key)
	err := r.client.ZAdd(ctx, fullKey, members...).Err()
	if err != nil {
		r.logger.Error("failed to zadd",
			zap.String("key", fullKey),
			zap.Error(err),
		)
		return err
	}
	return nil
}

// ZRange returns a range of members from a sorted set
func (r *RedisClient) ZRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	fullKey := r.makeKey(key)
	members, err := r.client.ZRange(ctx, fullKey, start, stop).Result()
	if err != nil {
		r.logger.Error("failed to zrange",
			zap.String("key", fullKey),
			zap.Error(err),
		)
		return nil, err
	}
	return members, nil
}

// ZRangeWithScores returns a range of members with scores from a sorted set
func (r *RedisClient) ZRangeWithScores(ctx context.Context, key string, start, stop int64) ([]redis.Z, error) {
	fullKey := r.makeKey(key)
	members, err := r.client.ZRangeWithScores(ctx, fullKey, start, stop).Result()
	if err != nil {
		r.logger.Error("failed to zrange with scores",
			zap.String("key", fullKey),
			zap.Error(err),
		)
		return nil, err
	}
	return members, nil
}

// Flush flushes all keys in the current database
func (r *RedisClient) Flush(ctx context.Context) error {
	err := r.client.FlushDB(ctx).Err()
	if err != nil {
		r.logger.Error("failed to flush database", zap.Error(err))
		return err
	}
	r.logger.Warn("flushed Redis database")
	return nil
}

// Ping checks if the Redis server is responsive
func (r *RedisClient) Ping(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

// Close closes the Redis connection
func (r *RedisClient) Close() error {
	r.logger.Info("closing Redis connection")
	return r.client.Close()
}

// GetClient returns the underlying Redis client
func (r *RedisClient) GetClient() *redis.Client {
	return r.client
}

// Stats returns connection pool statistics
func (r *RedisClient) Stats() *redis.PoolStats {
	return r.client.PoolStats()
}
