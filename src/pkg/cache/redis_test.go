package cache

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestRedis creates a miniredis server for testing
func setupTestRedis(t *testing.T) (*miniredis.Miniredis, *Redis) {
	mr, err := miniredis.Run()
	require.NoError(t, err, "Failed to start miniredis")

	config := Config{
		Addr:     mr.Addr(),
		Password: "",
		DB:       0,
		TTL:      5 * time.Minute,
	}

	redisCache := Connect(config)
	return mr, redisCache
}

func TestConnect(t *testing.T) {
	tests := []struct {
		name   string
		config Config
	}{
		{
			name: "successful connection with default config",
			config: Config{
				Addr:     "localhost:6379",
				Password: "",
				DB:       0,
				TTL:      5 * time.Minute,
			},
		},
		{
			name: "connection with custom TTL",
			config: Config{
				Addr:     "localhost:6379",
				Password: "",
				DB:       0,
				TTL:      10 * time.Minute,
			},
		},
		{
			name: "connection with password and different DB",
			config: Config{
				Addr:     "localhost:6379",
				Password: "testpassword",
				DB:       1,
				TTL:      3 * time.Minute,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			redisCache := Connect(tt.config)

			assert.NotNil(t, redisCache)
			assert.NotNil(t, redisCache.Client)
			assert.Equal(t, tt.config.TTL, redisCache.DefaultTTL)

			// Verify client options
			opts := redisCache.Client.Options()
			assert.Equal(t, tt.config.Addr, opts.Addr)
			assert.Equal(t, tt.config.Password, opts.Password)
			assert.Equal(t, tt.config.DB, opts.DB)
			assert.Equal(t, 5*time.Second, opts.DialTimeout)
			assert.Equal(t, 5*time.Second, opts.ReadTimeout)
			assert.Equal(t, 5*time.Second, opts.WriteTimeout)
			assert.Equal(t, 50, opts.PoolSize)
			assert.Equal(t, 10, opts.MinIdleConns)
		})
	}
}

func TestRedis_Health(t *testing.T) {
	t.Run("healthy connection", func(t *testing.T) {
		mr, redisCache := setupTestRedis(t)
		defer mr.Close()
		defer redisCache.Close()

		ctx := context.Background()
		err := redisCache.Health(ctx)

		assert.NoError(t, err)
	})

	t.Run("unhealthy connection - server stopped", func(t *testing.T) {
		mr, redisCache := setupTestRedis(t)
		mr.Close() // Close the server immediately
		defer redisCache.Close()

		ctx := context.Background()
		err := redisCache.Health(ctx)

		assert.Error(t, err)
	})

	t.Run("health check with timeout context", func(t *testing.T) {
		mr, redisCache := setupTestRedis(t)
		defer mr.Close()
		defer redisCache.Close()

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		err := redisCache.Health(ctx)
		assert.NoError(t, err)
	})

	t.Run("health check with cancelled context", func(t *testing.T) {
		mr, redisCache := setupTestRedis(t)
		defer mr.Close()
		defer redisCache.Close()

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		err := redisCache.Health(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context canceled")
	})
}

func TestRedis_Close(t *testing.T) {
	t.Run("close connection successfully", func(t *testing.T) {
		mr, redisCache := setupTestRedis(t)
		defer mr.Close()

		err := redisCache.Close()
		assert.NoError(t, err)

		// Verify connection is closed by trying to ping
		ctx := context.Background()
		err = redisCache.Health(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "client is closed")
	})

	t.Run("close already closed connection", func(t *testing.T) {
		mr, redisCache := setupTestRedis(t)
		defer mr.Close()

		// Close once
		err := redisCache.Close()
		assert.NoError(t, err)

		// Close again - should handle gracefully
		err = redisCache.Close()
		// Redis client returns "redis: client is closed" error
		assert.Error(t, err)
	})
}

func TestRedis_Integration(t *testing.T) {
	t.Run("full lifecycle test", func(t *testing.T) {
		mr, redisCache := setupTestRedis(t)
		defer mr.Close()
		defer redisCache.Close()

		ctx := context.Background()

		// Test health check
		err := redisCache.Health(ctx)
		assert.NoError(t, err)

		// Test basic operations to verify connection works
		err = redisCache.Client.Set(ctx, "test_key", "test_value", redisCache.DefaultTTL).Err()
		assert.NoError(t, err)

		val, err := redisCache.Client.Get(ctx, "test_key").Result()
		assert.NoError(t, err)
		assert.Equal(t, "test_value", val)

		// Test close
		err = redisCache.Close()
		assert.NoError(t, err)
	})
}

func TestRedis_ConnectionPool(t *testing.T) {
	t.Run("verify connection pool configuration", func(t *testing.T) {
		mr, redisCache := setupTestRedis(t)
		defer mr.Close()
		defer redisCache.Close()

		opts := redisCache.Client.Options()

		assert.Equal(t, 50, opts.PoolSize, "Pool size should be 50")
		assert.Equal(t, 10, opts.MinIdleConns, "Min idle connections should be 10")
	})

	t.Run("verify timeout configurations", func(t *testing.T) {
		mr, redisCache := setupTestRedis(t)
		defer mr.Close()
		defer redisCache.Close()

		opts := redisCache.Client.Options()

		assert.Equal(t, 5*time.Second, opts.DialTimeout, "Dial timeout should be 5 seconds")
		assert.Equal(t, 5*time.Second, opts.ReadTimeout, "Read timeout should be 5 seconds")
		assert.Equal(t, 5*time.Second, opts.WriteTimeout, "Write timeout should be 5 seconds")
	})
}

func TestRedis_DefaultTTL(t *testing.T) {
	tests := []struct {
		name        string
		ttl         time.Duration
		expectedTTL time.Duration
	}{
		{
			name:        "5 minute TTL",
			ttl:         5 * time.Minute,
			expectedTTL: 5 * time.Minute,
		},
		{
			name:        "1 hour TTL",
			ttl:         1 * time.Hour,
			expectedTTL: 1 * time.Hour,
		},
		{
			name:        "30 second TTL",
			ttl:         30 * time.Second,
			expectedTTL: 30 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mr, err := miniredis.Run()
			require.NoError(t, err)
			defer mr.Close()

			config := Config{
				Addr: mr.Addr(),
				TTL:  tt.ttl,
			}

			redisCache := Connect(config)
			defer redisCache.Close()

			assert.Equal(t, tt.expectedTTL, redisCache.DefaultTTL)
		})
	}
}

// Benchmark tests
func BenchmarkRedis_Health(b *testing.B) {
	mr, err := miniredis.Run()
	require.NoError(b, err)
	defer mr.Close()

	config := Config{
		Addr: mr.Addr(),
		TTL:  5 * time.Minute,
	}

	redisCache := Connect(config)
	defer redisCache.Close()

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = redisCache.Health(ctx)
	}
}

func BenchmarkRedis_Connect(b *testing.B) {
	mr, err := miniredis.Run()
	require.NoError(b, err)
	defer mr.Close()

	config := Config{
		Addr: mr.Addr(),
		TTL:  5 * time.Minute,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		redisCache := Connect(config)
		redisCache.Close()
	}
}
