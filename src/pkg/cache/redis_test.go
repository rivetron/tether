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
			defer func() { _ = redisCache.Close() }()

			assert.NotNil(t, redisCache)
			assert.NotNil(t, redisCache.Client)
			assert.Equal(t, tt.config.TTL, redisCache.DefaultTTL)

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
		defer func() { _ = redisCache.Close() }()

		ctx := context.Background()
		err := redisCache.Health(ctx)
		assert.NoError(t, err)
	})

	t.Run("unhealthy connection - server stopped", func(t *testing.T) {
		mr, redisCache := setupTestRedis(t)
		mr.Close()
		defer func() { _ = redisCache.Close() }()

		ctx := context.Background()
		err := redisCache.Health(ctx)
		assert.Error(t, err)
	})

	t.Run("health check with timeout context", func(t *testing.T) {
		mr, redisCache := setupTestRedis(t)
		defer mr.Close()
		defer func() { _ = redisCache.Close() }()

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		err := redisCache.Health(ctx)
		assert.NoError(t, err)
	})

	t.Run("health check with cancelled context", func(t *testing.T) {
		mr, redisCache := setupTestRedis(t)
		defer mr.Close()
		defer func() { _ = redisCache.Close() }()

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

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

		ctx := context.Background()
		err = redisCache.Health(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "client is closed")
	})

	t.Run("close already closed connection", func(t *testing.T) {
		mr, redisCache := setupTestRedis(t)
		defer mr.Close()

		err := redisCache.Close()
		assert.NoError(t, err)

		err = redisCache.Close()
		assert.Error(t, err)
	})
}

func TestRedis_Integration(t *testing.T) {
	t.Run("full lifecycle test", func(t *testing.T) {
		mr, redisCache := setupTestRedis(t)
		defer mr.Close()
		defer func() { _ = redisCache.Close() }()

		ctx := context.Background()

		err := redisCache.Health(ctx)
		assert.NoError(t, err)

		err = redisCache.Client.Set(ctx, "test_key", "test_value", redisCache.DefaultTTL).Err()
		assert.NoError(t, err)

		val, err := redisCache.Client.Get(ctx, "test_key").Result()
		assert.NoError(t, err)
		assert.Equal(t, "test_value", val)

		err = redisCache.Close()
		assert.NoError(t, err)
	})
}

func TestRedis_ConnectionPool(t *testing.T) {
	t.Run("verify connection pool configuration", func(t *testing.T) {
		mr, redisCache := setupTestRedis(t)
		defer mr.Close()
		defer func() { _ = redisCache.Close() }()

		opts := redisCache.Client.Options()
		assert.Equal(t, 50, opts.PoolSize)
		assert.Equal(t, 10, opts.MinIdleConns)
	})

	t.Run("verify timeout configurations", func(t *testing.T) {
		mr, redisCache := setupTestRedis(t)
		defer mr.Close()
		defer func() { _ = redisCache.Close() }()

		opts := redisCache.Client.Options()
		assert.Equal(t, 5*time.Second, opts.DialTimeout)
		assert.Equal(t, 5*time.Second, opts.ReadTimeout)
		assert.Equal(t, 5*time.Second, opts.WriteTimeout)
	})
}

func TestRedis_DefaultTTL(t *testing.T) {
	tests := []struct {
		name        string
		ttl         time.Duration
		expectedTTL time.Duration
	}{
		{"5 minute TTL", 5 * time.Minute, 5 * time.Minute},
		{"1 hour TTL", 1 * time.Hour, 1 * time.Hour},
		{"30 second TTL", 30 * time.Second, 30 * time.Second},
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
			defer func() { _ = redisCache.Close() }()

			assert.Equal(t, tt.expectedTTL, redisCache.DefaultTTL)
		})
	}
}

// Benchmarks
func BenchmarkRedis_Health(b *testing.B) {
	mr, err := miniredis.Run()
	require.NoError(b, err)
	defer mr.Close()

	config := Config{
		Addr: mr.Addr(),
		TTL:  5 * time.Minute,
	}

	redisCache := Connect(config)
	defer func() { _ = redisCache.Close() }()

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
		_ = redisCache.Close()
	}
}

func TestRedis_SetAndGet(t *testing.T) {
	mr, redisCache := setupTestRedis(t)
	defer mr.Close()
	defer redisCache.Close()

	ctx := context.Background()

	type payload struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	tests := []struct {
		name  string
		key   string
		value payload
	}{
		{"simple struct", "user:1", payload{"tejas", 26}},
		{"another struct", "user:2", payload{"john", 30}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := redisCache.Set(ctx, tt.key, tt.value, 0)
			require.NoError(t, err)

			var got payload
			err = redisCache.Get(ctx, tt.key, &got)
			require.NoError(t, err)

			assert.Equal(t, tt.value, got)
		})
	}
}

func TestRedis_Get_KeyNotFound(t *testing.T) {
	mr, redisCache := setupTestRedis(t)
	defer mr.Close()
	defer redisCache.Close()

	ctx := context.Background()

	var dest any
	err := redisCache.Get(ctx, "does_not_exist", &dest)
	assert.Error(t, err)
}

func TestRedis_Delete(t *testing.T) {
	mr, redisCache := setupTestRedis(t)
	defer mr.Close()
	defer redisCache.Close()

	ctx := context.Background()

	_ = redisCache.Set(ctx, "key1", "value", 0)

	err := redisCache.Delete(ctx, "key1")
	require.NoError(t, err)

	ok, err := redisCache.Exists(ctx, "key1")
	require.NoError(t, err)
	assert.False(t, ok)
}

func TestRedis_Exists(t *testing.T) {
	mr, redisCache := setupTestRedis(t)
	defer mr.Close()
	defer redisCache.Close()

	ctx := context.Background()

	_ = redisCache.Set(ctx, "abc", "hello", 0)

	ok, err := redisCache.Exists(ctx, "abc")
	require.NoError(t, err)
	assert.True(t, ok)

	ok, err = redisCache.Exists(ctx, "nope")
	require.NoError(t, err)
	assert.False(t, ok)
}

func TestRedis_IncrementFunctions(t *testing.T) {
	mr, redisCache := setupTestRedis(t)
	defer mr.Close()
	defer redisCache.Close()

	ctx := context.Background()

	val, err := redisCache.Increment(ctx, "counter")
	require.NoError(t, err)
	assert.Equal(t, int64(1), val)

	val, err = redisCache.IncrementBy(ctx, "counter", 5)
	require.NoError(t, err)
	assert.Equal(t, int64(6), val)
}

func TestRedis_SetNX(t *testing.T) {
	mr, redisCache := setupTestRedis(t)
	defer mr.Close()
	defer redisCache.Close()

	ctx := context.Background()

	ok, err := redisCache.SetNX(ctx, "unique:key", "value", 0)
	require.NoError(t, err)
	assert.True(t, ok)

	ok, err = redisCache.SetNX(ctx, "unique:key", "new", 0)
	require.NoError(t, err)
	assert.False(t, ok)
}

func TestRedis_BatchSetAndGet(t *testing.T) {
	mr, redisCache := setupTestRedis(t)
	defer mr.Close()
	defer redisCache.Close()

	ctx := context.Background()

	items := map[string]any{
		"user:1": map[string]any{"name": "tejas", "age": 26},
		"user:2": map[string]any{"name": "ram", "age": 32},
	}

	err := redisCache.BatchSet(ctx, items, 0)
	require.NoError(t, err)

	result, err := redisCache.BatchGet(ctx, []string{"user:1", "user:2"})
	require.NoError(t, err)

	assert.Len(t, result, 2)
	assert.Equal(t, "tejas", result["user:1"].(map[string]any)["name"])
	assert.Equal(t, float64(32), result["user:2"].(map[string]any)["age"])
}

func TestRedis_BatchGet_EmptyList(t *testing.T) {
	mr, redisCache := setupTestRedis(t)
	defer mr.Close()
	defer redisCache.Close()

	ctx := context.Background()

	result, err := redisCache.BatchGet(ctx, []string{})
	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestCacheKeyFunctions(t *testing.T) {
	tests := []struct {
		name     string
		keyFunc  func() string
		expected string
	}{
		{
			"CacheKey",
			func() string { return CacheKey("demo", "123") },
			"tether:demo:123",
		},
		{
			"LinkCacheKey",
			func() string { return LinkCacheKey("abc") },
			"tether:link:abc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.keyFunc()
			assert.Equal(t, tt.expected, got)
		})
	}
}
