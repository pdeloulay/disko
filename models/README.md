# Disko Models Package

This package contains the database models and MongoDB connection utilities for the Disko application.

## Overview

The models package implements:
- MongoDB v2 driver integration
- Data models for Board and Idea entities
- RICE scoring system
- Validation utilities
- Database connection management
- Performance-optimized indexes

## Models

### Board Model
- Represents a board with ideas organized in columns
- Contains metadata like name, description, public link
- Manages column and field visibility settings
- Tracks admin ownership via Clerk user ID

### Idea Model
- Represents individual ideas within boards
- Includes RICE scoring (Reach, Impact, Confidence, Effort)
- Supports drag-and-drop positioning
- Tracks feedback (thumbs up, emoji reactions)
- Manages workflow status and progress

### Supporting Types
- `RICEScore`: Scoring system for idea prioritization
- `EmojiReaction`: Feedback mechanism for public users
- `ColumnType`: Workflow columns (Parking, Now, Next, Later, Release, Won't Do)
- `IdeaStatus`: Idea lifecycle states (Draft, Active, Done, Archived)

## Database Connection

### Features
- Automatic connection management with timeouts
- Connection pooling and error handling
- Graceful disconnection
- Environment-based configuration

### Indexes
The package automatically creates optimized indexes:
- `boards.admin_id`: Fast board queries by admin
- `boards.public_link`: Unique public link access
- `ideas.board_id + position`: Efficient idea ordering
- `ideas.board_id + column`: Fast column queries
- `ideas.board_id + status`: Status filtering
- Text search index on idea content

## Validation

### Board Validation
- Required fields: name, publicLink, adminId
- Length constraints on text fields
- Column and field visibility validation
- Automatic timestamp management

### Idea Validation
- Required fields: boardId, oneLiner, description, valueStatement
- RICE score validation (R: 0-100%, I: 0-100%, C: 1/2/4/8, E: 0-100%)
- Column and status validation
- Position and feedback count validation

## Usage

### Database Connection
```go
import "disko-backend/models"

// Connect to database
err := models.ConnectDatabase()
if err != nil {
    log.Fatal("Database connection failed:", err)
}

// Get collections
boardsCollection := models.GetCollection(models.BoardsCollection)
ideasCollection := models.GetCollection(models.IdeasCollection)

// Disconnect when done
defer models.DisconnectDatabase()
```

### Model Validation
```go
board := &models.Board{
    Name: "My Board",
    PublicLink: "unique-link-123",
    AdminID: "clerk-user-id",
}

if errors := models.ValidateBoard(board); len(errors) > 0 {
    // Handle validation errors
    log.Printf("Validation errors: %v", errors)
}
```

## Environment Variables

Required environment variables:
- `MONGODB_URI`: MongoDB connection string
- `MONGODB_DATABASE`: Database name (defaults to "disko_board")

## Testing

Run tests with:
```bash
go test ./models -v
```

Tests cover:
- Model validation
- RICE score calculations
- Column and status validation
- Database connection utilities
- Error handling