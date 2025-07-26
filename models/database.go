package models

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// Database holds the MongoDB client and database instance
type Database struct {
	Client *mongo.Client
	DB     *mongo.Database
}

// Global database instance
var DB *Database

// ConnectDatabase initializes the MongoDB connection
func ConnectDatabase() error {
	mongoURI := os.Getenv("MONGODB_URI")
	if mongoURI == "" {
		return fmt.Errorf("MONGODB_URI environment variable is not set")
	}

	dbName := os.Getenv("MONGODB_DATABASE")
	if dbName == "" {
		dbName = "disko" // default database name
	}

	// Set client options
	clientOptions := options.Client().ApplyURI(mongoURI)

	// Set connection timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Connect to MongoDB
	client, err := mongo.Connect(clientOptions)
	if err != nil {
		return fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Test the connection
	err = client.Ping(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	// Initialize global database instance
	DB = &Database{
		Client: client,
		DB:     client.Database(dbName),
	}

	log.Printf("Successfully connected to MongoDB database: %s", dbName)

	// Set up indexes
	if err := setupIndexes(); err != nil {
		return fmt.Errorf("failed to setup database indexes: %w", err)
	}

	return nil
}

// DisconnectDatabase closes the MongoDB connection
func DisconnectDatabase() error {
	if DB == nil || DB.Client == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := DB.Client.Disconnect(ctx)
	if err != nil {
		return fmt.Errorf("failed to disconnect from MongoDB: %w", err)
	}

	log.Println("Successfully disconnected from MongoDB")
	return nil
}

// GetCollection returns a MongoDB collection
func GetCollection(collectionName string) *mongo.Collection {
	if DB == nil || DB.DB == nil {
		log.Fatal("Database not initialized. Call ConnectDatabase() first.")
	}
	return DB.DB.Collection(collectionName)
}

// Collection names constants
const (
	BoardsCollection = "boards"
	IdeasCollection  = "ideas"
)

// setupIndexes creates the necessary indexes for performance optimization
func setupIndexes() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Boards collection indexes
	boardsCollection := GetCollection(BoardsCollection)

	// Index on admin_id for efficient board queries by admin
	_, err := boardsCollection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "admin_id", Value: 1},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to create admin_id index on boards: %w", err)
	}

	// Unique index on public_link for efficient public board access
	_, err = boardsCollection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "public_link", Value: 1},
		},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		return fmt.Errorf("failed to create public_link index on boards: %w", err)
	}

	// Ideas collection indexes
	ideasCollection := GetCollection(IdeasCollection)

	// Compound index on board_id and position for efficient idea ordering
	_, err = ideasCollection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "board_id", Value: 1},
			{Key: "position", Value: 1},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to create board_id_position index on ideas: %w", err)
	}

	// Compound index on board_id and column for efficient column queries
	_, err = ideasCollection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "board_id", Value: 1},
			{Key: "column", Value: 1},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to create board_id_column index on ideas: %w", err)
	}

	// Compound index on board_id and status for efficient status filtering
	_, err = ideasCollection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "board_id", Value: 1},
			{Key: "status", Value: 1},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to create board_id_status index on ideas: %w", err)
	}

	// Text index for search functionality
	_, err = ideasCollection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "one_liner", Value: "text"},
			{Key: "description", Value: "text"},
			{Key: "value_statement", Value: "text"},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to create text search index on ideas: %w", err)
	}

	log.Println("Successfully created database indexes")
	return nil
}

// DatabaseError represents a database operation error
type DatabaseError struct {
	Operation string
	Err       error
}

func (e *DatabaseError) Error() string {
	return fmt.Sprintf("database %s error: %v", e.Operation, e.Err)
}

// NewDatabaseError creates a new database error
func NewDatabaseError(operation string, err error) *DatabaseError {
	return &DatabaseError{
		Operation: operation,
		Err:       err,
	}
}

// IsConnectionError checks if the error is a connection-related error
func IsConnectionError(err error) bool {
	if err == nil {
		return false
	}

	// Check for common connection error patterns
	errStr := err.Error()
	connectionErrors := []string{
		"connection refused",
		"no reachable servers",
		"context deadline exceeded",
		"network is unreachable",
		"connection reset by peer",
	}

	for _, connErr := range connectionErrors {
		if contains(errStr, connErr) {
			return true
		}
	}

	return false
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || (len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			containsAt(s, substr))))
}

func containsAt(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
