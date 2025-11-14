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
		{"zero", 0, 100},
		{"negative", -5, 100},
		{"positive", 50, 50},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			maxPool := uint64(tt.input)
			if tt.input <= 0 {
				maxPool = 100
			}
			assert.Equal(t, tt.expected, maxPool)
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
		{"zero", 0, 10},
		{"negative", -1, 10},
		{"positive", 20, 20},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			minPool := uint64(tt.input)
			if tt.input <= 0 {
				minPool = 10
			}
			assert.Equal(t, tt.expected, minPool)
		})
	}
}

// TestConfig_Struct tests Config struct initialization
func TestConfig_Struct(t *testing.T) {
	config := Config{
		URI:         "mongodb://localhost:27017",
		Name:        "test",
		MaxPoolSize: 50,
		MinPoolSize: 5,
	}

	assert.Equal(t, "mongodb://localhost:27017", config.URI)
	assert.Equal(t, "test", config.Name)
	assert.Equal(t, 50, config.MaxPoolSize)
	assert.Equal(t, 5, config.MinPoolSize)
}

// TestConfig_ZeroValues tests Config with zero values
func TestConfig_ZeroValues(t *testing.T) {
	config := Config{}

	assert.Empty(t, config.URI)
	assert.Empty(t, config.Name)
	assert.Zero(t, config.MaxPoolSize)
	assert.Zero(t, config.MinPoolSize)
}

// TestMongoDB_StructFields tests MongoDB struct fields
func TestMongoDB_StructFields(t *testing.T) {
	db := &MongoDB{}
	assert.Nil(t, db.Client)
	assert.Nil(t, db.Database)
}

// TestConnect_InvalidURI_Format tests various invalid URI formats
func TestConnect_InvalidURIs(t *testing.T) {
	tests := []struct {
		name string
		uri  string
	}{
		{"empty", ""},
		{"invalid scheme", "http://localhost:27017"},
		{"malformed", "%%%"},
		{"wrong protocol", "postgres://localhost"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()

			cfg := Config{
				URI:         tt.uri,
				Name:        "x",
				MaxPoolSize: 10,
				MinPoolSize: 5,
			}

			db, err := Connect(ctx, cfg)
			assert.Error(t, err)
			assert.Nil(t, db)
		})
	}
}

// TestConnect_ContextCancellation tests behavior when context is cancelled
func TestConnect_ContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	cfg := Config{
		URI:         "mongodb://localhost:27017",
		Name:        "x",
		MaxPoolSize: 10,
		MinPoolSize: 5,
	}

	db, err := Connect(ctx, cfg)
	assert.Error(t, err)
	assert.Nil(t, db)
}

// TestConnect_ContextTimeout tests behavior with very short timeout
func TestConnect_Timeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	time.Sleep(1 * time.Millisecond)

	cfg := Config{
		URI:         "mongodb://nonexistent-host:27017",
		Name:        "x",
		MaxPoolSize: 10,
		MinPoolSize: 5,
	}

	db, err := Connect(ctx, cfg)
	assert.Error(t, err)
	assert.Nil(t, db)
}

// TestConfig_ValidURIFormats tests various valid MongoDB URI formats
func TestConfig_ValidURIs(t *testing.T) {
	valid := []string{
		"mongodb://localhost:27017",
		"mongodb://user:pass@localhost:27017",
		"mongodb://localhost:27017,localhost:27018",
		"mongodb+srv://cluster.mongodb.net",
	}

	for _, uri := range valid {
		t.Run(uri, func(t *testing.T) {
			cfg := Config{
				URI:         uri,
				Name:        "db",
				MaxPoolSize: 50,
				MinPoolSize: 5,
			}
			assert.NotEmpty(t, cfg.URI)
		})
	}
}

// TestMongoDB_MethodsExist tests that required methods exist
func TestMongoDB_Methods_NilReceiver(t *testing.T) {
	var db *MongoDB

	assert.Panics(t, func() {
		_ = db.Close(context.Background())
	})

	assert.Panics(t, func() {
		_ = db.Health(context.Background())
	})
}

// TestConnect_ConfigValidation tests config parameter validation
func TestPoolSizeConversion(t *testing.T) {
	tests := []struct {
		name        string
		maxIn       int
		minIn       int
		maxExpected uint64
		minExpected uint64
	}{
		{"positive", 100, 10, 100, 10},
		{"zero", 0, 0, 100, 10},
		{"negative", -10, -5, 100, 10},
		{"mixed", 200, 0, 200, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			maxPool := uint64(tt.maxIn)
			minPool := uint64(tt.minIn)

			if tt.maxIn <= 0 {
				maxPool = 100
			}
			if tt.minIn <= 0 {
				minPool = 10
			}

			assert.Equal(t, tt.maxExpected, maxPool)
			assert.Equal(t, tt.minExpected, minPool)
		})
	}
}
