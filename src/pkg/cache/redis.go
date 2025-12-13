package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

type Redis struct {
	Client     *redis.Client
	DefaultTTL time.Duration
}

type Config struct {
	Addr     string
	Password string
	DB       int
	TTL      time.Duration
}

func Connect(config Config) *Redis {
	db := redis.NewClient(&redis.Options{
		Addr:         config.Addr,
		Password:     config.Password,
		DB:           config.DB,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		PoolSize:     50,
		MinIdleConns: 10,
	})

	return &Redis{
		Client:     db,
		DefaultTTL: config.TTL,
	}
}

func (r *Redis) Close() error {
	return r.Client.Close()
}

func (r *Redis) Health(ctx context.Context) error {
	return r.Client.Ping(ctx).Err()
}

// <-------------------- Main Functions -------------------->

func (r *Redis) GetClient() *redis.Client {
	return r.Client
}

func (r *Redis) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	if ttl == 0 {
		ttl = r.DefaultTTL
	}

	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	return r.Client.Set(ctx, key, data, ttl).Err()
}

func (r *Redis) Get(ctx context.Context, key string, dest any) error {
	data, err := r.Client.Get(ctx, key).Result()
	if err != nil {
		return fmt.Errorf("failed to get value: %w", err)
	}

	if err := json.Unmarshal([]byte(data), dest); err != nil {
		return fmt.Errorf("failed to unmarshal value :%w", err)
	}

	return nil
}

func (r *Redis) Delete(ctx context.Context, key string) error {
	return r.Client.Del(ctx, key).Err()
}

func (r *Redis) Exists(ctx context.Context, key string) (bool, error) {
	count, err := r.Client.Exists(ctx, key).Result()
	return count > 0, err
}

func (r *Redis) Increment(ctx context.Context, key string) (int64, error) {
	return r.Client.Incr(ctx, key).Result()
}

func (r *Redis) IncrementBy(ctx context.Context, key string, value int64) (int64, error) {
	return r.Client.IncrBy(ctx, key, value).Result()
}

func (r *Redis) SetNX(ctx context.Context, key string, value any, ttl time.Duration) (bool, error) {
	if ttl == 0 {
		ttl = r.DefaultTTL
	}

	data, err := json.Marshal(value)
	if err != nil {
		return false, fmt.Errorf("failed to marshal value: %w", err)
	}

	return r.Client.SetNX(ctx, key, data, ttl).Result()
}

// <-------------------- Batch Functions -------------------->

func (r *Redis) BatchGet(ctx context.Context, keys []string) (map[string]any, error) {
	if len(keys) == 0 {
		return make(map[string]any), nil
	}

	pipe := r.Client.Pipeline()
	cmds := make(map[string]*redis.StringCmd)

	for _, key := range keys {
		cmds[key] = pipe.Get(ctx, key)
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to execute pipeline: %w", err)
	}

	result := make(map[string]any)
	for key, cmd := range cmds {
		if cmd.Err() == nil {
			var value any
			if err := json.Unmarshal([]byte(cmd.Val()), &value); err == nil {
				result[key] = value
			}

		}
	}

	return result, nil
}

func (r *Redis) BatchSet(ctx context.Context, items map[string]any, ttl time.Duration) error {
	if len(items) == 0 {
		return nil
	}

	if ttl == 0 {
		ttl = r.DefaultTTL
	}

	pipe := r.Client.Pipeline()

	for key, value := range items {
		data, err := json.Marshal(value)
		if err != nil {
			return fmt.Errorf("failed to marshal value for key %s: %w", key, err)
		}

		pipe.Set(ctx, key, data, ttl)
	}

	_, err := pipe.Exec(ctx)
	return err
}

// <-------------------- Key Functions -------------------->
func CacheKey(prefix, key string) string {
	return fmt.Sprintf("tether:%s:%s", prefix, key)
}

// LinkCacheKey generates cache key for links
func LinkCacheKey(shortCode string) string {
	return CacheKey("link", shortCode)
}
