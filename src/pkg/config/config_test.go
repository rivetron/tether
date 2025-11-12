package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad_WithDefaults(t *testing.T) {
	// Clear any existing environment variables
	clearEnv(t)

	config, err := Load()
	require.NoError(t, err)
	require.NotNil(t, config)

	// Test Server defaults
	assert.Equal(t, 5000, config.Server.Port)
	assert.Equal(t, "0.0.0.0", config.Server.Host)
	assert.Equal(t, "http://localhost:5000", config.Server.BaseURL)
	assert.Equal(t, 30*time.Second, config.Server.ReadTimeout)
	assert.Equal(t, 30*time.Second, config.Server.WriteTimeout)
	assert.Equal(t, 120*time.Second, config.Server.IdleTimeout)
	assert.Equal(t, "debug", config.Server.GinMode)

	// Test App defaults
	assert.Equal(t, 720*time.Hour, config.App.DefaultTTL)
	assert.Equal(t, 8760*time.Hour, config.App.MaxTTL)
	assert.Equal(t, 7, config.App.ShortCodeLen)
	assert.Equal(t, 50, config.App.CustomCodeLen)
	assert.Equal(t, 50, config.App.MaxCustomCode)
	assert.True(t, config.App.EnableStats)
	assert.True(t, config.App.EnableGeoIP)

	// Test Database defaults
	assert.Equal(t, "mongodb://localhost:27017", config.Database.URI)
	assert.Equal(t, "tether", config.Database.Name)
	assert.Equal(t, 100, config.Database.MaxPoolSize)
	assert.Equal(t, 10, config.Database.MinPoolSize)

	// Test Redis (Cache) defaults
	assert.Equal(t, "localhost:6379", config.Cache.Addr)
	assert.Equal(t, "", config.Cache.Password)
	assert.Equal(t, 0, config.Cache.DB)
	assert.Equal(t, 24*time.Hour, config.Cache.TTL)

	// Test Security defaults
	assert.Equal(t, []string{"*"}, config.Security.AllowedOrigins)
	assert.Empty(t, config.Security.TrustedProxies)
	assert.True(t, config.Security.EnableCors)
	assert.Empty(t, config.Security.BlockedDomains)
	assert.Empty(t, config.Security.RequiredHeaders)

	// Test Logging defaults
	assert.Equal(t, "info", config.Logging.Level)
	assert.Equal(t, "json", config.Logging.Format)
	assert.Equal(t, "stdout", config.Logging.Output)
}

func TestLoad_WithEnvironmentVariables(t *testing.T) {
	clearEnv(t)

	// Set environment variables
	os.Setenv("SERVER_PORT", "9090")
	os.Setenv("SERVER_HOST", "127.0.0.1")
	os.Setenv("BASE_URL", "https://example.com")
	os.Setenv("GIN_MODE", "release")
	os.Setenv("MONGO_URI", "mongodb://testdb:27017")
	os.Setenv("MONGO_DATABASE", "testdb")
	os.Setenv("REDIS_ADDR", "redis:6379")
	os.Setenv("REDIS_PASSWORD", "secret")
	os.Setenv("REDIS_DB", "1")
	os.Setenv("LOG_LEVEL", "debug")

	defer clearEnv(t)

	config, err := Load()
	require.NoError(t, err)
	require.NotNil(t, config)

	// Verify environment variables are loaded
	assert.Equal(t, 9090, config.Server.Port)
	assert.Equal(t, "127.0.0.1", config.Server.Host)
	assert.Equal(t, "https://example.com", config.Server.BaseURL)
	assert.Equal(t, "release", config.Server.GinMode)
	assert.Equal(t, "mongodb://testdb:27017", config.Database.URI)
	assert.Equal(t, "testdb", config.Database.Name)
	assert.Equal(t, "redis:6379", config.Cache.Addr)
	assert.Equal(t, "secret", config.Cache.Password)
	assert.Equal(t, 1, config.Cache.DB)
	assert.Equal(t, "debug", config.Logging.Level)
}

func TestLoad_ServerConfig(t *testing.T) {
	tests := []struct {
		name     string
		envVars  map[string]string
		validate func(t *testing.T, config *Config)
	}{
		{
			name: "custom port",
			envVars: map[string]string{
				"SERVER_PORT": "3000",
			},
			validate: func(t *testing.T, config *Config) {
				assert.Equal(t, 3000, config.Server.Port)
			},
		},
		{
			name: "production mode",
			envVars: map[string]string{
				"GIN_MODE": "release",
			},
			validate: func(t *testing.T, config *Config) {
				assert.Equal(t, "release", config.Server.GinMode)
			},
		},
		{
			name: "custom base URL",
			envVars: map[string]string{
				"BASE_URL": "https://myapp.com",
			},
			validate: func(t *testing.T, config *Config) {
				assert.Equal(t, "https://myapp.com", config.Server.BaseURL)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearEnv(t)
			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}
			defer clearEnv(t)

			config, err := Load()
			require.NoError(t, err)
			tt.validate(t, config)
		})
	}
}

func TestLoad_DatabaseConfig(t *testing.T) {
	tests := []struct {
		name     string
		envVars  map[string]string
		validate func(t *testing.T, config *Config)
	}{
		{
			name: "custom MongoDB URI",
			envVars: map[string]string{
				"MONGO_URI": "mongodb://user:pass@mongo:27017",
			},
			validate: func(t *testing.T, config *Config) {
				assert.Equal(t, "mongodb://user:pass@mongo:27017", config.Database.URI)
			},
		},
		{
			name: "custom database name",
			envVars: map[string]string{
				"MONGO_DATABASE": "production_db",
			},
			validate: func(t *testing.T, config *Config) {
				assert.Equal(t, "production_db", config.Database.Name)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearEnv(t)
			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}
			defer clearEnv(t)

			config, err := Load()
			require.NoError(t, err)
			tt.validate(t, config)
		})
	}
}

func TestLoad_CacheConfig(t *testing.T) {
	tests := []struct {
		name     string
		envVars  map[string]string
		validate func(t *testing.T, config *Config)
	}{
		{
			name: "Redis with password",
			envVars: map[string]string{
				"REDIS_ADDR":     "redis.example.com:6379",
				"REDIS_PASSWORD": "supersecret",
				"REDIS_DB":       "2",
			},
			validate: func(t *testing.T, config *Config) {
				assert.Equal(t, "redis.example.com:6379", config.Cache.Addr)
				assert.Equal(t, "supersecret", config.Cache.Password)
				assert.Equal(t, 2, config.Cache.DB)
			},
		},
		{
			name: "Redis without password",
			envVars: map[string]string{
				"REDIS_ADDR": "localhost:6380",
			},
			validate: func(t *testing.T, config *Config) {
				assert.Equal(t, "localhost:6380", config.Cache.Addr)
				assert.Equal(t, "", config.Cache.Password)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearEnv(t)
			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}
			defer clearEnv(t)

			config, err := Load()
			require.NoError(t, err)
			tt.validate(t, config)
		})
	}
}

func TestLoad_LoggingConfig(t *testing.T) {
	tests := []struct {
		name     string
		envVars  map[string]string
		validate func(t *testing.T, config *Config)
	}{
		{
			name: "debug logging",
			envVars: map[string]string{
				"LOG_LEVEL": "debug",
			},
			validate: func(t *testing.T, config *Config) {
				assert.Equal(t, "debug", config.Logging.Level)
			},
		},
		{
			name: "error logging",
			envVars: map[string]string{
				"LOG_LEVEL": "error",
			},
			validate: func(t *testing.T, config *Config) {
				assert.Equal(t, "error", config.Logging.Level)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearEnv(t)
			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}
			defer clearEnv(t)

			config, err := Load()
			require.NoError(t, err)
			tt.validate(t, config)
		})
	}
}

func TestLoad_AppConfig(t *testing.T) {
	clearEnv(t)
	defer clearEnv(t)

	config, err := Load()
	require.NoError(t, err)

	// Test that App config has sensible defaults
	assert.Equal(t, 720*time.Hour, config.App.DefaultTTL, "Default TTL should be 30 days")
	assert.Equal(t, 8760*time.Hour, config.App.MaxTTL, "Max TTL should be 1 year")
	assert.Greater(t, config.App.ShortCodeLen, 0, "Short code length should be positive")
	assert.Greater(t, config.App.CustomCodeLen, 0, "Custom code length should be positive")
}

func TestLoad_SecurityConfig(t *testing.T) {
	clearEnv(t)
	defer clearEnv(t)

	config, err := Load()
	require.NoError(t, err)

	// Test Security config defaults
	assert.True(t, config.Security.EnableCors, "CORS should be enabled by default")
	assert.NotEmpty(t, config.Security.AllowedOrigins, "Allowed origins should have default")
}

func TestLoad_DurationParsing(t *testing.T) {
	clearEnv(t)
	defer clearEnv(t)

	config, err := Load()
	require.NoError(t, err)

	// Test that durations are properly parsed
	assert.IsType(t, time.Duration(0), config.Server.ReadTimeout)
	assert.IsType(t, time.Duration(0), config.Server.WriteTimeout)
	assert.IsType(t, time.Duration(0), config.Server.IdleTimeout)
	assert.IsType(t, time.Duration(0), config.Cache.TTL)
	assert.IsType(t, time.Duration(0), config.App.DefaultTTL)
	assert.IsType(t, time.Duration(0), config.App.MaxTTL)
}

func TestLoad_MultipleCallsReturnDifferentInstances(t *testing.T) {
	clearEnv(t)
	defer clearEnv(t)

	config1, err1 := Load()
	require.NoError(t, err1)

	config2, err2 := Load()
	require.NoError(t, err2)

	// Configs should have same values but be different instances
	assert.Equal(t, config1.Server.Port, config2.Server.Port)
	assert.NotSame(t, config1, config2, "Load should return new instances")
}

// Helper function to clear all relevant environment variables
func clearEnv(t *testing.T) {
	t.Helper()
	envVars := []string{
		"SERVER_PORT",
		"SERVER_HOST",
		"BASE_URL",
		"GIN_MODE",
		"MONGO_URI",
		"MONGO_DATABASE",
		"REDIS_ADDR",
		"REDIS_PASSWORD",
		"REDIS_DB",
		"LOG_LEVEL",
	}

	for _, v := range envVars {
		os.Unsetenv(v)
	}
}
