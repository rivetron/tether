package database

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoDB struct {
	Client   *mongo.Client
	Database *mongo.Database
}

type Config struct {
	URI         string
	Name        string
	MaxPoolSize int
	MinPoolSize int
}

func Connect(ctx context.Context, config Config) (*MongoDB, error) {
	// * Validate pool sizes to prevent overflow
	maxPoolSize := uint64(config.MaxPoolSize)
	minPoolSize := uint64(config.MinPoolSize)
	if config.MaxPoolSize <= 0 {
		maxPoolSize = 100
	}
	if config.MinPoolSize <= 0 {
		minPoolSize = 10
	}

	clientOptions := options.Client().
		ApplyURI(config.URI).
		SetMaxPoolSize(maxPoolSize).
		SetMinPoolSize(minPoolSize).
		SetMaxConnIdleTime(30 * time.Minute).
		SetServerSelectionTimeout(5 * time.Second).
		SetConnectTimeout(10 * time.Second)

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDb: %w", err)
	}

	// ? Ping the DB to check connection
	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	database := client.Database(config.Name)

	return &MongoDB{
		Client:   client,
		Database: database,
	}, nil
}

func (m *MongoDB) Close(ctx context.Context) error {
	return m.Client.Disconnect(ctx)
}

func (m *MongoDB) Health(ctx context.Context) error {
	return m.Client.Ping(ctx, nil)
}
