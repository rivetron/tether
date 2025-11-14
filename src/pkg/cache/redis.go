package cache

import (
	"context"
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
