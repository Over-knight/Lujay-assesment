package storage

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// MongoDB holds the MongoDB client and database instance
type MongoDB struct {
	Client   *mongo.Client
	Database *mongo.Database
}

// MongoConfig holds MongoDB connection configuration
type MongoConfig struct {
	URI      string
	Database string
	Timeout  time.Duration
}

// NewMongoDB creates a new MongoDB connection
// It establishes a connection to MongoDB and verifies it with a ping
// Returns a MongoDB instance or an error if connection fails
func NewMongoDB(config MongoConfig) (*MongoDB, error) {
	// Set default timeout if not provided
	if config.Timeout == 0 {
		config.Timeout = 10 * time.Second
	}

	// Create context with timeout for connection
	ctx, cancel := context.WithTimeout(context.Background(), config.Timeout)
	defer cancel()

	// Set client options
	clientOptions := options.Client().ApplyURI(config.URI)

	// Connect to MongoDB
	log.Println("Connecting to MongoDB...")
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Ping the database to verify connection
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	log.Println("Successfully connected to MongoDB")

	// Return MongoDB instance
	return &MongoDB{
		Client:   client,
		Database: client.Database(config.Database),
	}, nil
}

// Close closes the MongoDB connection
// Should be called when the application shuts down
func (m *MongoDB) Close(ctx context.Context) error {
	log.Println("Closing MongoDB connection...")
	if err := m.Client.Disconnect(ctx); err != nil {
		return fmt.Errorf("failed to disconnect from MongoDB: %w", err)
	}
	log.Println("MongoDB connection closed")
	return nil
}

// Collection returns a handle to a specific collection
// This is a helper method to access collections from the database
func (m *MongoDB) Collection(name string) *mongo.Collection {
	return m.Database.Collection(name)
}

// Ping checks if the MongoDB connection is alive
// Returns an error if the connection is not working
func (m *MongoDB) Ping(ctx context.Context) error {
	return m.Client.Ping(ctx, readpref.Primary())
}
