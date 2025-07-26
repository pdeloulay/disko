package models

import (
	"os"
	"testing"
)

func TestDatabaseConnection(t *testing.T) {
	// Skip this test if no MongoDB URI is provided
	mongoURI := os.Getenv("MONGODB_URI")
	if mongoURI == "" {
		t.Skip("Skipping database connection test - MONGODB_URI not set")
	}

	// Test connection
	err := ConnectDatabase()
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	// Test that we can get collections
	boardsCollection := GetCollection(BoardsCollection)
	if boardsCollection == nil {
		t.Error("Failed to get boards collection")
	}

	ideasCollection := GetCollection(IdeasCollection)
	if ideasCollection == nil {
		t.Error("Failed to get ideas collection")
	}

	// Test disconnection
	err = DisconnectDatabase()
	if err != nil {
		t.Errorf("Failed to disconnect from database: %v", err)
	}
}

func TestIsConnectionError(t *testing.T) {
	// Test nil error
	if IsConnectionError(nil) {
		t.Error("Expected nil error to not be a connection error")
	}

	// Test connection-related error messages
	connectionErrors := []string{
		"connection refused",
		"no reachable servers",
		"context deadline exceeded",
		"network is unreachable",
		"connection reset by peer",
	}

	for _, errMsg := range connectionErrors {
		// Create a mock error with the message
		mockErr := &DatabaseError{
			Operation: "test",
			Err:       &mockError{message: errMsg},
		}

		if !IsConnectionError(mockErr) {
			t.Errorf("Expected error with message '%s' to be identified as connection error", errMsg)
		}
	}

	// Test non-connection error
	nonConnErr := &DatabaseError{
		Operation: "test",
		Err:       &mockError{message: "validation failed"},
	}

	if IsConnectionError(nonConnErr) {
		t.Error("Expected validation error to not be identified as connection error")
	}
}

// mockError is a simple error implementation for testing
type mockError struct {
	message string
}

func (e *mockError) Error() string {
	return e.message
}

func TestDatabaseError(t *testing.T) {
	originalErr := &mockError{message: "original error"}
	dbErr := NewDatabaseError("insert", originalErr)

	expectedMsg := "database insert error: original error"
	if dbErr.Error() != expectedMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedMsg, dbErr.Error())
	}

	if dbErr.Operation != "insert" {
		t.Errorf("Expected operation 'insert', got '%s'", dbErr.Operation)
	}

	if dbErr.Err != originalErr {
		t.Error("Expected wrapped error to match original error")
	}
}
