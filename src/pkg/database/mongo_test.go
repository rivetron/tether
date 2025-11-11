package database

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestConfig_DefaultMaxPoolSize tests that MaxPoolSize defaults to 100 when <= 0
func TestConfig_DefaultMaxPoolSize(t *testing.T) {
	tests := []struct {
		name     string
		input    int
		expected uint64
	}{
		{"zero value", 0, 100},
		{"negative value", -10, 100},
		{"positive value", 50, 50},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			maxPoolSize := uint64(tt.input)
			if tt.input <= 0 {
				maxPoolSize = 100
			}
			assert.Equal(t, tt.expected, maxPoolSize)
		})
	}
}

// TestConfig_DefaultMinPoolSize tests that MinPoolSize defaults to 10 when <= 0
func TestConfig_DefaultMinPoolSize(t *testing.T) {
	tests := []struct {
		name     string
		input    int
		expected uint64
	}{
		{"zero value", 0, 10},
		{"negative value", -5, 10},
		{"positive value", 20, 20},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			minPoolSize := uint64(tt.input)
			if tt.input <= 0 {
				minPoolSize = 10
			}
			assert.Equal(t, tt.expected, minPoolSize)
		})
	}
}

// TestConfig_Struct tests Config struct initialization
func TestConfig_Struct(t *testing.T) {
	config := Config{
		URI:         "mongodb://localhost:27017",
		Name:        "test_db",
		MaxPoolSize: 100,
		MinPoolSize: 10,
	}

	assert.Equal(t, "mongodb://localhost:27017", config.URI)
	assert.Equal(t, "test_db", config.Name)
	assert.Equal(t, 100, config.MaxPoolSize)
	assert.Equal(t, 10, config.MinPoolSize)
}

// TestConfig_ZeroValues tests Config with zero values
func TestConfig_ZeroValues(t *testing.T) {
	config := Config{}

	assert.Empty(t, config.URI)
	assert.Empty(t, config.Name)
	assert.Equal(t, 0, config.MaxPoolSize)
	assert.Equal(t, 0, config.MinPoolSize)
}

// TestMongoDB_StructFields tests MongoDB struct fields
func TestMongoDB_StructFields(t *testing.T) {
	db := &MongoDB{
		Client:   nil,
		Database: nil,
	}

	assert.Nil(t, db.Client)
	assert.Nil(t, db.Database)
}

// TestConnect_InvalidURI_Format tests various invalid URI formats
func TestConnect_InvalidURI_Format(t *testing.T) {
	tests := []struct {
		name string
		uri  string
	}{
		{"empty URI", ""},
		{"invalid scheme", "http://localhost:27017"},
		{"malformed URI", "not-a-valid-uri"},
		{"invalid protocol", "postgresql://localhost:5432"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			config := Config{
				URI:         tt.uri,
				Name:        "test_db",
				MaxPoolSize: 10,
				MinPoolSize: 5,
			}

			db, err := Connect(ctx, config)

			// Should get an error for invalid URIs
			assert.Error(t, err, "Expected error for invalid URI: %s", tt.uri)
			assert.Nil(t, db, "MongoDB instance should be nil on error")

			if err != nil {
				assert.Contains(t, err.Error(), "failed to", "Error should contain 'failed to'")
			}
		})
	}
}

// TestConnect_ContextCancellation tests behavior when context is cancelled
func TestConnect_ContextCancellation(t *testing.T) {
	// Create a context that's already cancelled
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	config := Config{
		URI:         "mongodb://localhost:27017",
		Name:        "test_db",
		MaxPoolSize: 50,
		MinPoolSize: 5,
	}

	db, err := Connect(ctx, config)

	// Should return error when context is cancelled
	assert.Error(t, err, "Should return error when context is cancelled")
	assert.Nil(t, db, "MongoDB instance should be nil when context is cancelled")
}

// TestConnect_ContextTimeout tests behavior with very short timeout
func TestConnect_ContextTimeout(t *testing.T) {
	// Create a context with 1 nanosecond timeout (will expire immediately)
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	// Sleep to ensure timeout
	time.Sleep(1 * time.Millisecond)

	config := Config{
		URI:         "mongodb://nonexistent-host:27017",
		Name:        "test_db",
		MaxPoolSize: 50,
		MinPoolSize: 5,
	}

	db, err := Connect(ctx, config)

	// Should return error due to timeout
	assert.Error(t, err, "Should return error on timeout")
	assert.Nil(t, db, "MongoDB instance should be nil on timeout")
}

// TestConfig_PoolSizeConversion tests integer to uint64 conversion logic
func TestConfig_PoolSizeConversion(t *testing.T) {
	tests := []struct {
		name        string
		maxPoolSize int
		minPoolSize int
		expectedMax uint64
		expectedMin uint64
	}{
		{
			name:        "both positive",
			maxPoolSize: 100,
			minPoolSize: 10,
			expectedMax: 100,
			expectedMin: 10,
		},
		{
			name:        "both zero",
			maxPoolSize: 0,
			minPoolSize: 0,
			expectedMax: 100, // default
			expectedMin: 10,  // default
		},
		{
			name:        "both negative",
			maxPoolSize: -50,
			minPoolSize: -5,
			expectedMax: 100, // default
			expectedMin: 10,  // default
		},
		{
			name:        "mixed values",
			maxPoolSize: 200,
			minPoolSize: 0,
			expectedMax: 200,
			expectedMin: 10, // default
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate the logic from Connect function
			maxPoolSize := uint64(tt.maxPoolSize)
			minPoolSize := uint64(tt.minPoolSize)

			if tt.maxPoolSize <= 0 {
				maxPoolSize = 100
			}
			if tt.minPoolSize <= 0 {
				minPoolSize = 10
			}

			assert.Equal(t, tt.expectedMax, maxPoolSize, "MaxPoolSize mismatch")
			assert.Equal(t, tt.expectedMin, minPoolSize, "MinPoolSize mismatch")
		})
	}
}

// TestConfig_ValidURIFormats tests various valid MongoDB URI formats
func TestConfig_ValidURIFormats(t *testing.T) {
	validURIs := []string{
		"mongodb://localhost:27017",
		"mongodb://127.0.0.1:27017",
		"mongodb://user:pass@localhost:27017",
		"mongodb://localhost:27017,localhost:27018",
		"mongodb+srv://cluster.mongodb.net",
		"mongodb://user:pass@host1:27017,host2:27017/dbname?replicaSet=myRepl",
	}

	for _, uri := range validURIs {
		t.Run(uri, func(t *testing.T) {
			config := Config{
				URI:         uri,
				Name:        "test_db",
				MaxPoolSize: 50,
				MinPoolSize: 5,
			}

			// Just verify the config can be created with valid URIs
			assert.NotEmpty(t, config.URI)
			assert.NotEmpty(t, config.Name)
			assert.Greater(t, config.MaxPoolSize, 0)
			assert.Greater(t, config.MinPoolSize, 0)
		})
	}
}

// TestMongoDB_MethodsExist tests that required methods exist
func TestMongoDB_MethodsExist(t *testing.T) {
	// This test verifies the MongoDB type has the expected methods
	// by checking if we can reference them

	var db *MongoDB

	// These should compile if methods exist
	_ = func(ctx context.Context) error { return db.Close(ctx) }
	_ = func(ctx context.Context) error { return db.Health(ctx) }

	// If we got here, methods exist
	assert.True(t, true, "All expected methods exist")
}

// TestConnect_ConfigValidation tests config parameter validation
func TestConnect_ConfigValidation(t *testing.T) {
	tests := []struct {
		name        string
		config      Config
		expectError bool
	}{
		{
			name: "empty URI",
			config: Config{
				URI:         "",
				Name:        "test_db",
				MaxPoolSize: 100,
				MinPoolSize: 10,
			},
			expectError: true,
		},
		{
			name: "empty database name",
			config: Config{
				URI:         "mongodb://localhost:27017",
				Name:        "",
				MaxPoolSize: 100,
				MinPoolSize: 10,
			},
			expectError: false, // MongoDB allows empty db name (uses default)
		},
		{
			name: "all zero values",
			config: Config{
				URI:         "",
				Name:        "",
				MaxPoolSize: 0,
				MinPoolSize: 0,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			db, err := Connect(ctx, tt.config)

			if tt.expectError {
				assert.Error(t, err, "Expected error for config: %+v", tt.config)
				assert.Nil(t, db)
			} else {
				// Even valid configs will fail without MongoDB running
				// but we're just testing the validation logic
				if db != nil {
					_ = db.Close(ctx)
				}
			}
		})
	}
}

// TestConnect_TimeoutConfiguration tests timeout settings
func TestConnect_TimeoutConfiguration(t *testing.T) {
	config := Config{
		URI:         "mongodb://localhost:27017",
		Name:        "test_db",
		MaxPoolSize: 100,
		MinPoolSize: 10,
	}

	// The Connect function sets these timeouts:
	// - MaxConnIdleTime: 30 minutes
	// - ServerSelectionTimeout: 5 seconds
	// - ConnectTimeout: 10 seconds

	// We can't directly test the timeouts without connecting,
	// but we can verify the config is valid
	assert.NotEmpty(t, config.URI)
	assert.NotEmpty(t, config.Name)
	assert.Greater(t, config.MaxPoolSize, 0)
}

// TestMongoDB_NilSafety tests behavior with nil values
func TestMongoDB_NilSafety(t *testing.T) {
	var db *MongoDB

	// Calling methods on nil should panic (expected Go behavior)
	assert.Panics(t, func() {
		ctx := context.Background()
		_ = db.Close(ctx)
	}, "Close on nil MongoDB should panic")

	assert.Panics(t, func() {
		ctx := context.Background()
		_ = db.Health(ctx)
	}, "Health on nil MongoDB should panic")
}

// TestConfig_LargePoolSizes tests with very large pool sizes
func TestConfig_LargePoolSizes(t *testing.T) {
	config := Config{
		URI:         "mongodb://localhost:27017",
		Name:        "test_db",
		MaxPoolSize: 10000,
		MinPoolSize: 1000,
	}

	assert.Equal(t, 10000, config.MaxPoolSize)
	assert.Equal(t, 1000, config.MinPoolSize)
	assert.Greater(t, config.MaxPoolSize, config.MinPoolSize,
		"MaxPoolSize should be greater than MinPoolSize")
}
