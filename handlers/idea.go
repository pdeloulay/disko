package handlers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"disko-backend/middleware"
	"disko-backend/models"
	"disko-backend/utils"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// CreateIdeaRequest represents the request payload for creating an idea
type CreateIdeaRequest struct {
	OneLiner       string           `json:"oneLiner" binding:"required,min=1,max=200"`
	Description    string           `json:"description" binding:"required,min=1,max=1000"`
	ValueStatement string           `json:"valueStatement" binding:"required,min=1,max=500"`
	RiceScore      models.RICEScore `json:"riceScore" binding:"required"`
	Column         string           `json:"column,omitempty"`
	Position       int              `json:"position,omitempty"`
}

// UpdateIdeaRequest represents the request payload for updating an idea
type UpdateIdeaRequest struct {
	OneLiner       string            `json:"oneLiner,omitempty" binding:"omitempty,min=1,max=200"`
	Description    string            `json:"description,omitempty" binding:"omitempty,min=1,max=1000"`
	ValueStatement string            `json:"valueStatement,omitempty" binding:"omitempty,min=1,max=500"`
	RiceScore      *models.RICEScore `json:"riceScore,omitempty"`
	Column         string            `json:"column,omitempty"`
	InProgress     *bool             `json:"inProgress,omitempty"`
	Status         string            `json:"status,omitempty"`
}

// UpdateIdeaPositionRequest represents the request payload for updating idea position
type UpdateIdeaPositionRequest struct {
	Column   string `json:"column" binding:"required"`
	Position int    `json:"position" binding:"min=0"`
}

// UpdateIdeaStatusRequest represents the request payload for updating idea status
type UpdateIdeaStatusRequest struct {
	InProgress *bool  `json:"inProgress,omitempty"`
	Status     string `json:"status,omitempty"`
	Column     string `json:"column,omitempty"`
}

// IdeaResponse represents the response format for idea operations
type IdeaResponse struct {
	ID             string                 `json:"id"`
	BoardID        string                 `json:"boardId"`
	OneLiner       string                 `json:"oneLiner"`
	Description    string                 `json:"description"`
	ValueStatement string                 `json:"valueStatement"`
	RiceScore      models.RICEScore       `json:"riceScore"`
	Column         string                 `json:"column"`
	Position       int                    `json:"position"`
	InProgress     bool                   `json:"inProgress"`
	Status         string                 `json:"status"`
	ThumbsUp       int                    `json:"thumbsUp"`
	EmojiReactions []models.EmojiReaction `json:"emojiReactions"`
	CreatedAt      time.Time              `json:"createdAt"`
	UpdatedAt      time.Time              `json:"updatedAt"`
}

// PublicIdeaResponse represents the response format for public idea access (filtered)
type PublicIdeaResponse struct {
	ID             string                 `json:"id"`
	OneLiner       string                 `json:"oneLiner"`
	Description    string                 `json:"description,omitempty"`
	ValueStatement string                 `json:"valueStatement,omitempty"`
	Column         string                 `json:"column"`
	Position       int                    `json:"position"`
	InProgress     bool                   `json:"inProgress"`
	ThumbsUp       int                    `json:"thumbsUp"`
	EmojiReactions []models.EmojiReaction `json:"emojiReactions"`
	CreatedAt      time.Time              `json:"createdAt"`
	UpdatedAt      time.Time              `json:"updatedAt"`
}

// CreateIdea handles POST /api/boards/:id/ideas
func CreateIdea(c *gin.Context) {
	// Get user ID from auth middleware
	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "INTERNAL_ERROR",
				"message": "Failed to get user ID",
			},
		})
		return
	}

	// Get board ID from URL parameter
	boardID := c.Param("id")
	if boardID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "INVALID_BOARD_ID",
				"message": "Board ID is required",
			},
		})
		return
	}

	// Parse request body
	var req CreateIdeaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "VALIDATION_ERROR",
				"message": "Invalid request data",
				"details": err.Error(),
			},
		})
		return
	}

	// Validate RICE score
	if !req.RiceScore.IsValidRICEScore() {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "INVALID_RICE_SCORE",
				"message": "Invalid RICE score values. R: 0-100%, I: 0-100%, C: 1/2/4/8, E: 0-100%",
			},
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Verify board exists and belongs to user
	boardsCollection := models.GetCollection(models.BoardsCollection)
	boardFilter := bson.M{
		"_id":     boardID,
		"user_id": userID,
	}

	var board models.Board
	err = boardsCollection.FindOne(ctx, boardFilter).Decode(&board)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{
				"error": gin.H{
					"code":    "BOARD_NOT_FOUND",
					"message": "Board not found or you don't have permission to add ideas",
				},
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "DATABASE_ERROR",
				"message": "Failed to verify board",
				"details": err.Error(),
			},
		})
		return
	}

	// Set default column to parking if not specified
	column := req.Column
	if column == "" {
		column = string(models.ColumnParking)
	}

	// Validate column
	if !models.IsValidColumn(column) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "INVALID_COLUMN",
				"message": "Invalid column type: " + column,
			},
		})
		return
	}

	// Get next position in column if not specified
	position := req.Position
	if position == 0 {
		ideasCollection := models.GetCollection(models.IdeasCollection)
		positionFilter := bson.M{
			"board_id": boardID,
			"column":   column,
		}

		// Find the highest position in the column
		opts := options.FindOne().SetSort(bson.D{{Key: "position", Value: -1}})
		var lastIdea models.Idea
		err = ideasCollection.FindOne(ctx, positionFilter, opts).Decode(&lastIdea)
		if err != nil && err != mongo.ErrNoDocuments {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": gin.H{
					"code":    "DATABASE_ERROR",
					"message": "Failed to determine position",
					"details": err.Error(),
				},
			})
			return
		}

		if err == mongo.ErrNoDocuments {
			position = 1 // First idea in column
		} else {
			position = lastIdea.Position + 1
		}
	}

	// Generate unique idea ID with "I" prefix
	ideaID := utils.GenerateIdeaID()

	// Create idea document
	now := time.Now().UTC()
	idea := models.Idea{
		ID:             ideaID,
		BoardID:        boardID,
		OneLiner:       req.OneLiner,
		Description:    req.Description,
		ValueStatement: req.ValueStatement,
		RiceScore:      req.RiceScore,
		Column:         column,
		Position:       position,
		InProgress:     false,
		Status:         string(models.StatusActive),
		ThumbsUp:       0,
		EmojiReactions: []models.EmojiReaction{},
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	// Validate idea
	if validationErrors := models.ValidateIdea(&idea); len(validationErrors) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "VALIDATION_ERROR",
				"message": "Idea validation failed",
				"details": validationErrors.Error(),
			},
		})
		return
	}

	// Insert into MongoDB
	ideasCollection := models.GetCollection(models.IdeasCollection)
	_, err = ideasCollection.InsertOne(ctx, idea)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "DATABASE_ERROR",
				"message": "Failed to create idea",
				"details": err.Error(),
			},
		})
		return
	}

	// Return created idea
	response := IdeaResponse{
		ID:             idea.ID,
		BoardID:        idea.BoardID,
		OneLiner:       idea.OneLiner,
		Description:    idea.Description,
		ValueStatement: idea.ValueStatement,
		RiceScore:      idea.RiceScore,
		Column:         idea.Column,
		Position:       idea.Position,
		InProgress:     idea.InProgress,
		Status:         idea.Status,
		ThumbsUp:       idea.ThumbsUp,
		EmojiReactions: idea.EmojiReactions,
		CreatedAt:      idea.CreatedAt,
		UpdatedAt:      idea.UpdatedAt,
	}

	c.JSON(http.StatusCreated, response)
}

// GetBoardIdeas handles GET /api/boards/:id/ideas
func GetBoardIdeas(c *gin.Context) {
	startTime := time.Now()
	boardID := c.Param("id")
	userAgent := c.GetHeader("User-Agent")
	referer := c.GetHeader("Referer")

	// Get user ID from auth middleware
	userID, err := middleware.GetUserID(c)
	if err != nil {
		log.Printf("[Handler] GetBoardIdeas failed - GetUserID error: %v, BoardID: %s, IP: %s, UserAgent: %s", err, boardID, c.ClientIP(), userAgent)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "INTERNAL_ERROR",
				"message": "Failed to get user ID",
			},
		})
		return
	}

	log.Printf("[Handler] GetBoardIdeas started - BoardID: %s, UserID: %s, IP: %s, UserAgent: %s, Referer: %s",
		boardID, userID, c.ClientIP(), userAgent, referer)
	log.Printf("[Handler] GetBoardIdeas - Request headers: %+v", c.Request.Header)
	log.Printf("[Handler] GetBoardIdeas - Authorization header: %s", c.GetHeader("Authorization"))

	// Get board ID from URL parameter
	if boardID == "" {
		log.Printf("[Handler] GetBoardIdeas failed - Empty board ID, UserID: %s", userID)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "INVALID_BOARD_ID",
				"message": "Board ID is required",
			},
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Verify board exists and belongs to user
	boardsCollection := models.GetCollection(models.BoardsCollection)
	boardFilter := bson.M{
		"_id":     boardID,
		"user_id": userID,
	}

	log.Printf("[Handler] GetBoardIdeas - Verifying board ownership: Filter: %+v, BoardID: %s, UserID: %s", boardFilter, boardID, userID)

	var board models.Board
	err = boardsCollection.FindOne(ctx, boardFilter).Decode(&board)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{
				"error": gin.H{
					"code":    "BOARD_NOT_FOUND",
					"message": "Board not found or you don't have permission to view ideas",
				},
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "DATABASE_ERROR",
				"message": "Failed to verify board",
				"details": err.Error(),
			},
		})
		return
	}

	// Query ideas for the board
	ideasCollection := models.GetCollection(models.IdeasCollection)
	ideasFilter := bson.M{"board_id": boardID}

	log.Printf("[Handler] GetBoardIdeas - Querying ideas: Filter: %+v, BoardID: %s", ideasFilter, boardID)

	// Sort by column and position
	opts := options.Find().SetSort(bson.D{
		{Key: "column", Value: 1},
		{Key: "position", Value: 1},
	})

	cursor, err := ideasCollection.Find(ctx, ideasFilter, opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "DATABASE_ERROR",
				"message": "Failed to fetch ideas",
				"details": err.Error(),
			},
		})
		return
	}
	defer cursor.Close(ctx)

	// Decode results
	var ideas []models.Idea
	if err := cursor.All(ctx, &ideas); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "DATABASE_ERROR",
				"message": "Failed to decode ideas",
				"details": err.Error(),
			},
		})
		return
	}

	// Convert to response format
	var responses []IdeaResponse
	for _, idea := range ideas {
		responses = append(responses, IdeaResponse{
			ID:             idea.ID,
			BoardID:        idea.BoardID,
			OneLiner:       idea.OneLiner,
			Description:    idea.Description,
			ValueStatement: idea.ValueStatement,
			RiceScore:      idea.RiceScore,
			Column:         idea.Column,
			Position:       idea.Position,
			InProgress:     idea.InProgress,
			Status:         idea.Status,
			ThumbsUp:       idea.ThumbsUp,
			EmojiReactions: idea.EmojiReactions,
			CreatedAt:      idea.CreatedAt,
			UpdatedAt:      idea.UpdatedAt,
		})
	}

	duration := time.Since(startTime)
	log.Printf("[Handler] GetBoardIdeas success - BoardID: %s, UserID: %s, Ideas count: %d, Duration: %v, IP: %s",
		boardID, userID, len(responses), duration, c.ClientIP())

	c.JSON(http.StatusOK, gin.H{
		"ideas": responses,
		"count": len(responses),
	})
}

// UpdateIdea handles PUT /api/ideas/:id
func UpdateIdea(c *gin.Context) {
	// Get user ID from auth middleware
	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "INTERNAL_ERROR",
				"message": "Failed to get user ID",
			},
		})
		return
	}

	// Get idea ID from URL parameter
	ideaID := c.Param("id")
	if ideaID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "INVALID_IDEA_ID",
				"message": "Idea ID is required",
			},
		})
		return
	}

	// Parse request body
	var req UpdateIdeaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "VALIDATION_ERROR",
				"message": "Invalid request data",
				"details": err.Error(),
			},
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// First, get the idea to verify it exists and get board info
	ideasCollection := models.GetCollection(models.IdeasCollection)
	var existingIdea models.Idea
	err = ideasCollection.FindOne(ctx, bson.M{"_id": ideaID}).Decode(&existingIdea)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{
				"error": gin.H{
					"code":    "IDEA_NOT_FOUND",
					"message": "Idea not found",
				},
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "DATABASE_ERROR",
				"message": "Failed to fetch idea",
				"details": err.Error(),
			},
		})
		return
	}

	// Verify user owns the board containing this idea
	boardsCollection := models.GetCollection(models.BoardsCollection)
	boardFilter := bson.M{
		"_id":     existingIdea.BoardID,
		"user_id": userID,
	}

	var board models.Board
	err = boardsCollection.FindOne(ctx, boardFilter).Decode(&board)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusForbidden, gin.H{
				"error": gin.H{
					"code":    "PERMISSION_DENIED",
					"message": "You don't have permission to update this idea",
				},
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "DATABASE_ERROR",
				"message": "Failed to verify board ownership",
				"details": err.Error(),
			},
		})
		return
	}

	// Build update document
	updateDoc := bson.M{
		"updated_at": time.Now().UTC(),
	}

	if req.OneLiner != "" {
		updateDoc["one_liner"] = req.OneLiner
	}

	if req.Description != "" {
		updateDoc["description"] = req.Description
	}

	if req.ValueStatement != "" {
		updateDoc["value_statement"] = req.ValueStatement
	}

	if req.RiceScore != nil {
		// Validate RICE score
		if !req.RiceScore.IsValidRICEScore() {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": gin.H{
					"code":    "INVALID_RICE_SCORE",
					"message": "Invalid RICE score values. R: 0-100%, I: 0-100%, C: 1/2/4/8, E: 0-100%",
				},
			})
			return
		}
		updateDoc["rice_score"] = req.RiceScore
	}

	if req.Column != "" {
		// Validate column
		if !models.IsValidColumn(req.Column) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": gin.H{
					"code":    "INVALID_COLUMN",
					"message": "Invalid column type: " + req.Column,
				},
			})
			return
		}
		updateDoc["column"] = req.Column
	}

	if req.InProgress != nil {
		updateDoc["in_progress"] = *req.InProgress
	}

	if req.Status != "" {
		// Validate status
		if !models.IsValidStatus(req.Status) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": gin.H{
					"code":    "INVALID_STATUS",
					"message": "Invalid status: " + req.Status,
				},
			})
			return
		}
		updateDoc["status"] = req.Status

		// Automatic column transitions based on status
		switch req.Status {
		case string(models.StatusDone):
			// When marked as done, move to release column and remove in-progress
			updateDoc["column"] = string(models.ColumnRelease)
			updateDoc["in_progress"] = false
		case string(models.StatusArchived):
			// When archived, move to wont-do column and remove in-progress
			updateDoc["column"] = string(models.ColumnWontDo)
			updateDoc["in_progress"] = false
		case string(models.StatusActive):
			// When reactivated, move back to parking if currently in release or wont-do
			if existingIdea.Column == string(models.ColumnRelease) || existingIdea.Column == string(models.ColumnWontDo) {
				updateDoc["column"] = string(models.ColumnParking)
			}
		}
	}

	// Update idea in MongoDB
	filter := bson.M{"_id": ideaID}
	result, err := ideasCollection.UpdateOne(ctx, filter, bson.M{"$set": updateDoc})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "DATABASE_ERROR",
				"message": "Failed to update idea",
				"details": err.Error(),
			},
		})
		return
	}

	if result.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": gin.H{
				"code":    "IDEA_NOT_FOUND",
				"message": "Idea not found",
			},
		})
		return
	}

	// Fetch and return updated idea
	var updatedIdea models.Idea
	err = ideasCollection.FindOne(ctx, filter).Decode(&updatedIdea)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "DATABASE_ERROR",
				"message": "Failed to fetch updated idea",
				"details": err.Error(),
			},
		})
		return
	}

	// Return updated idea
	response := IdeaResponse{
		ID:             updatedIdea.ID,
		BoardID:        updatedIdea.BoardID,
		OneLiner:       updatedIdea.OneLiner,
		Description:    updatedIdea.Description,
		ValueStatement: updatedIdea.ValueStatement,
		RiceScore:      updatedIdea.RiceScore,
		Column:         updatedIdea.Column,
		Position:       updatedIdea.Position,
		InProgress:     updatedIdea.InProgress,
		Status:         updatedIdea.Status,
		ThumbsUp:       updatedIdea.ThumbsUp,
		EmojiReactions: updatedIdea.EmojiReactions,
		CreatedAt:      updatedIdea.CreatedAt,
		UpdatedAt:      updatedIdea.UpdatedAt,
	}

	c.JSON(http.StatusOK, response)
}

// DeleteIdea handles DELETE /api/ideas/:id
func DeleteIdea(c *gin.Context) {
	// Get user ID from auth middleware
	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "INTERNAL_ERROR",
				"message": "Failed to get user ID",
			},
		})
		return
	}

	// Get idea ID from URL parameter
	ideaID := c.Param("id")
	if ideaID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "INVALID_IDEA_ID",
				"message": "Idea ID is required",
			},
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// First, get the idea to verify it exists and get board info
	ideasCollection := models.GetCollection(models.IdeasCollection)
	var existingIdea models.Idea
	err = ideasCollection.FindOne(ctx, bson.M{"_id": ideaID}).Decode(&existingIdea)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{
				"error": gin.H{
					"code":    "IDEA_NOT_FOUND",
					"message": "Idea not found",
				},
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "DATABASE_ERROR",
				"message": "Failed to fetch idea",
				"details": err.Error(),
			},
		})
		return
	}

	// Verify user owns the board containing this idea
	boardsCollection := models.GetCollection(models.BoardsCollection)
	boardFilter := bson.M{
		"_id":     existingIdea.BoardID,
		"user_id": userID,
	}

	var board models.Board
	err = boardsCollection.FindOne(ctx, boardFilter).Decode(&board)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusForbidden, gin.H{
				"error": gin.H{
					"code":    "PERMISSION_DENIED",
					"message": "You don't have permission to delete this idea",
				},
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "DATABASE_ERROR",
				"message": "Failed to verify board ownership",
				"details": err.Error(),
			},
		})
		return
	}

	// Delete the idea
	filter := bson.M{"_id": ideaID}
	result, err := ideasCollection.DeleteOne(ctx, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "DATABASE_ERROR",
				"message": "Failed to delete idea",
				"details": err.Error(),
			},
		})
		return
	}

	if result.DeletedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": gin.H{
				"code":    "IDEA_NOT_FOUND",
				"message": "Idea not found",
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Idea deleted successfully",
	})
}

// UpdateIdeaPosition handles PUT /api/ideas/:id/position
func UpdateIdeaPosition(c *gin.Context) {
	// Get user ID from auth middleware
	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "INTERNAL_ERROR",
				"message": "Failed to get user ID",
			},
		})
		return
	}

	// Get idea ID from URL parameter
	ideaID := c.Param("id")
	if ideaID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "INVALID_IDEA_ID",
				"message": "Idea ID is required",
			},
		})
		return
	}

	// Parse request body
	var req UpdateIdeaPositionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "VALIDATION_ERROR",
				"message": "Invalid request data",
				"details": err.Error(),
			},
		})
		return
	}

	// Validate column
	if !models.IsValidColumn(req.Column) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "INVALID_COLUMN",
				"message": "Invalid column type: " + req.Column,
			},
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// First, get the idea to verify it exists and get board info
	ideasCollection := models.GetCollection(models.IdeasCollection)
	var existingIdea models.Idea
	err = ideasCollection.FindOne(ctx, bson.M{"_id": ideaID}).Decode(&existingIdea)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{
				"error": gin.H{
					"code":    "IDEA_NOT_FOUND",
					"message": "Idea not found",
				},
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "DATABASE_ERROR",
				"message": "Failed to fetch idea",
				"details": err.Error(),
			},
		})
		return
	}

	// Verify user owns the board containing this idea
	boardsCollection := models.GetCollection(models.BoardsCollection)
	boardFilter := bson.M{
		"_id":     existingIdea.BoardID,
		"user_id": userID,
	}

	var board models.Board
	err = boardsCollection.FindOne(ctx, boardFilter).Decode(&board)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusForbidden, gin.H{
				"error": gin.H{
					"code":    "PERMISSION_DENIED",
					"message": "You don't have permission to update this idea",
				},
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "DATABASE_ERROR",
				"message": "Failed to verify board ownership",
				"details": err.Error(),
			},
		})
		return
	}

	// Update idea position and column
	updateDoc := bson.M{
		"column":     req.Column,
		"position":   req.Position,
		"updated_at": time.Now().UTC(),
	}

	// If moving back to parking, remove in-progress status
	if req.Column == string(models.ColumnParking) {
		updateDoc["in_progress"] = false
	}

	filter := bson.M{"_id": ideaID}
	result, err := ideasCollection.UpdateOne(ctx, filter, bson.M{"$set": updateDoc})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "DATABASE_ERROR",
				"message": "Failed to update idea position",
				"details": err.Error(),
			},
		})
		return
	}

	if result.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": gin.H{
				"code":    "IDEA_NOT_FOUND",
				"message": "Idea not found",
			},
		})
		return
	}

	// Fetch and return updated idea
	var updatedIdea models.Idea
	err = ideasCollection.FindOne(ctx, filter).Decode(&updatedIdea)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "DATABASE_ERROR",
				"message": "Failed to fetch updated idea",
				"details": err.Error(),
			},
		})
		return
	}

	// Return updated idea
	response := IdeaResponse{
		ID:             updatedIdea.ID,
		BoardID:        updatedIdea.BoardID,
		OneLiner:       updatedIdea.OneLiner,
		Description:    updatedIdea.Description,
		ValueStatement: updatedIdea.ValueStatement,
		RiceScore:      updatedIdea.RiceScore,
		Column:         updatedIdea.Column,
		Position:       updatedIdea.Position,
		InProgress:     updatedIdea.InProgress,
		Status:         updatedIdea.Status,
		ThumbsUp:       updatedIdea.ThumbsUp,
		EmojiReactions: updatedIdea.EmojiReactions,
		CreatedAt:      updatedIdea.CreatedAt,
		UpdatedAt:      updatedIdea.UpdatedAt,
	}

	// Broadcast idea position update to WebSocket clients
	positionUpdate := map[string]interface{}{
		"ideaId":   ideaID,
		"column":   req.Column,
		"position": req.Position,
		"type":     "position_update",
	}
	utils.BroadcastIdeaUpdate(updatedIdea.BoardID, ideaID, positionUpdate)

	c.JSON(http.StatusOK, response)
}

// UpdateIdeaStatus handles PUT /api/ideas/:id/status
func UpdateIdeaStatus(c *gin.Context) {
	// Get user ID from auth middleware
	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "INTERNAL_ERROR",
				"message": "Failed to get user ID",
			},
		})
		return
	}

	// Get idea ID from URL parameter
	ideaID := c.Param("id")
	if ideaID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "INVALID_IDEA_ID",
				"message": "Idea ID is required",
			},
		})
		return
	}

	// Parse request body
	var req UpdateIdeaStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "VALIDATION_ERROR",
				"message": "Invalid request data",
				"details": err.Error(),
			},
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// First, get the idea to verify it exists and get board info
	ideasCollection := models.GetCollection(models.IdeasCollection)
	var existingIdea models.Idea
	err = ideasCollection.FindOne(ctx, bson.M{"_id": ideaID}).Decode(&existingIdea)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{
				"error": gin.H{
					"code":    "IDEA_NOT_FOUND",
					"message": "Idea not found",
				},
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "DATABASE_ERROR",
				"message": "Failed to fetch idea",
				"details": err.Error(),
			},
		})
		return
	}

	// Verify user owns the board containing this idea
	boardsCollection := models.GetCollection(models.BoardsCollection)
	boardFilter := bson.M{
		"_id":     existingIdea.BoardID,
		"user_id": userID,
	}

	var board models.Board
	err = boardsCollection.FindOne(ctx, boardFilter).Decode(&board)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusForbidden, gin.H{
				"error": gin.H{
					"code":    "PERMISSION_DENIED",
					"message": "You don't have permission to update this idea",
				},
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "DATABASE_ERROR",
				"message": "Failed to verify board ownership",
				"details": err.Error(),
			},
		})
		return
	}

	// Build update document
	updateDoc := bson.M{
		"updated_at": time.Now().UTC(),
	}

	// Handle in-progress status update
	if req.InProgress != nil {
		updateDoc["in_progress"] = *req.InProgress
	}

	// Handle status update with automatic column transitions
	if req.Status != "" {
		// Validate status
		if !models.IsValidStatus(req.Status) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": gin.H{
					"code":    "INVALID_STATUS",
					"message": "Invalid status: " + req.Status,
				},
			})
			return
		}

		updateDoc["status"] = req.Status

		// Automatic column transitions based on status
		switch req.Status {
		case string(models.StatusDone):
			// When marked as done, move to release column and remove in-progress
			updateDoc["column"] = string(models.ColumnRelease)
			updateDoc["in_progress"] = false
		case string(models.StatusArchived):
			// When archived, move to wont-do column and remove in-progress
			updateDoc["column"] = string(models.ColumnWontDo)
			updateDoc["in_progress"] = false
		case string(models.StatusActive):
			// When reactivated, move back to parking if currently in release or wont-do
			if existingIdea.Column == string(models.ColumnRelease) || existingIdea.Column == string(models.ColumnWontDo) {
				updateDoc["column"] = string(models.ColumnParking)
			}
		}
	}

	// Handle explicit column update (overrides automatic transitions)
	if req.Column != "" {
		// Validate column
		if !models.IsValidColumn(req.Column) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": gin.H{
					"code":    "INVALID_COLUMN",
					"message": "Invalid column type: " + req.Column,
				},
			})
			return
		}
		updateDoc["column"] = req.Column

		// If moving back to parking, remove in-progress status
		if req.Column == string(models.ColumnParking) {
			updateDoc["in_progress"] = false
		}
	}

	// Update idea in MongoDB
	filter := bson.M{"_id": ideaID}
	result, err := ideasCollection.UpdateOne(ctx, filter, bson.M{"$set": updateDoc})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "DATABASE_ERROR",
				"message": "Failed to update idea status",
				"details": err.Error(),
			},
		})
		return
	}

	if result.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": gin.H{
				"code":    "IDEA_NOT_FOUND",
				"message": "Idea not found",
			},
		})
		return
	}

	// Fetch and return updated idea
	var updatedIdea models.Idea
	err = ideasCollection.FindOne(ctx, filter).Decode(&updatedIdea)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "DATABASE_ERROR",
				"message": "Failed to fetch updated idea",
				"details": err.Error(),
			},
		})
		return
	}

	// Return updated idea
	response := IdeaResponse{
		ID:             updatedIdea.ID,
		BoardID:        updatedIdea.BoardID,
		OneLiner:       updatedIdea.OneLiner,
		Description:    updatedIdea.Description,
		ValueStatement: updatedIdea.ValueStatement,
		RiceScore:      updatedIdea.RiceScore,
		Column:         updatedIdea.Column,
		Position:       updatedIdea.Position,
		InProgress:     updatedIdea.InProgress,
		Status:         updatedIdea.Status,
		ThumbsUp:       updatedIdea.ThumbsUp,
		EmojiReactions: updatedIdea.EmojiReactions,
		CreatedAt:      updatedIdea.CreatedAt,
		UpdatedAt:      updatedIdea.UpdatedAt,
	}

	// Broadcast idea status update to WebSocket clients
	statusUpdate := map[string]interface{}{
		"ideaId":     ideaID,
		"inProgress": updatedIdea.InProgress,
		"status":     updatedIdea.Status,
		"column":     updatedIdea.Column,
		"type":       "status_update",
	}
	utils.BroadcastIdeaUpdate(updatedIdea.BoardID, ideaID, statusUpdate)

	c.JSON(http.StatusOK, response)
}

// GetPublicBoardIdeas handles GET /api/boards/:id/ideas/public
func GetPublicBoardIdeas(c *gin.Context) {
	// Get public link from URL parameter
	publicLink := c.Param("id")
	if publicLink == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "INVALID_PUBLIC_LINK",
				"message": "Public link is required",
			},
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// First, find the board by public link and ensure it's public
	boardsCollection := models.GetCollection(models.BoardsCollection)
	boardFilter := bson.M{"public_link": publicLink, "is_public": true}

	var board models.Board
	err := boardsCollection.FindOne(ctx, boardFilter).Decode(&board)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{
				"error": gin.H{
					"code":    "BOARD_NOT_FOUND",
					"message": "Board not found or is not publicly accessible. The board owner must make it public first.",
				},
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "DATABASE_ERROR",
				"message": "Failed to fetch board",
				"details": err.Error(),
			},
		})
		return
	}

	// Query ideas for the board
	ideasCollection := models.GetCollection(models.IdeasCollection)
	ideasFilter := bson.M{"board_id": board.ID}

	// Sort by column and position
	opts := options.Find().SetSort(bson.D{
		{Key: "column", Value: 1},
		{Key: "position", Value: 1},
	})

	cursor, err := ideasCollection.Find(ctx, ideasFilter, opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "DATABASE_ERROR",
				"message": "Failed to fetch ideas",
				"details": err.Error(),
			},
		})
		return
	}
	defer cursor.Close(ctx)

	// Decode results
	var ideas []models.Idea
	if err := cursor.All(ctx, &ideas); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "DATABASE_ERROR",
				"message": "Failed to decode ideas",
				"details": err.Error(),
			},
		})
		return
	}

	// Filter ideas based on visible columns
	visibleColumns := make(map[string]bool)
	for _, column := range board.VisibleColumns {
		visibleColumns[column] = true
	}

	// Filter visible fields
	visibleFields := make(map[string]bool)
	for _, field := range board.VisibleFields {
		visibleFields[field] = true
	}

	// Convert to public response format with field filtering
	var responses []PublicIdeaResponse
	for _, idea := range ideas {
		// Only include ideas in visible columns
		if !visibleColumns[idea.Column] {
			continue
		}

		response := PublicIdeaResponse{
			ID:             idea.ID,
			OneLiner:       idea.OneLiner, // Always visible
			Column:         idea.Column,
			Position:       idea.Position,
			InProgress:     idea.InProgress,
			ThumbsUp:       idea.ThumbsUp,
			EmojiReactions: idea.EmojiReactions,
			CreatedAt:      idea.CreatedAt,
			UpdatedAt:      idea.UpdatedAt,
		}

		// Add optional fields based on visibility settings
		if visibleFields[string(models.FieldDescription)] {
			response.Description = idea.Description
		}

		if visibleFields[string(models.FieldValueStatement)] {
			response.ValueStatement = idea.ValueStatement
		}

		// Note: RICE scores are never included in public view for privacy

		responses = append(responses, response)
	}

	c.JSON(http.StatusOK, gin.H{
		"ideas": responses,
		"count": len(responses),
		"board": gin.H{
			"id":             board.ID,
			"name":           board.Name,
			"description":    board.Description,
			"visibleColumns": board.VisibleColumns,
			"visibleFields":  board.VisibleFields,
		},
	})
}

// ThumbsUpRequest represents the request for thumbs up feedback
type ThumbsUpRequest struct {
	// No body needed - just IP-based rate limiting
}

// EmojiReactionRequest represents the request for emoji feedback
type EmojiReactionRequest struct {
	Emoji string `json:"emoji" binding:"required,min=1,max=10"`
}

// AddThumbsUp handles POST /api/ideas/:id/thumbsup (public endpoint)
func AddThumbsUp(c *gin.Context) {
	// Get idea ID from URL parameter
	ideaID := c.Param("id")
	if ideaID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "INVALID_IDEA_ID",
				"message": "Idea ID is required",
			},
		})
		return
	}

	// Get client IP for rate limiting
	clientIP := c.ClientIP()

	// Simple rate limiting: check if this IP has made a request in the last 5 seconds
	// In production, you'd use Redis or similar for distributed rate limiting
	rateLimitKey := "thumbsup_" + ideaID + "_" + clientIP
	if isRateLimited(rateLimitKey, 5*time.Second) {
		c.JSON(http.StatusTooManyRequests, gin.H{
			"error": gin.H{
				"code":    "RATE_LIMITED",
				"message": "Please wait before giving another thumbs up",
			},
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Find the idea and verify it exists
	ideasCollection := models.GetCollection(models.IdeasCollection)
	var idea models.Idea
	err := ideasCollection.FindOne(ctx, bson.M{"_id": ideaID}).Decode(&idea)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{
				"error": gin.H{
					"code":    "IDEA_NOT_FOUND",
					"message": "Idea not found",
				},
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "DATABASE_ERROR",
				"message": "Failed to fetch idea",
				"details": err.Error(),
			},
		})
		return
	}

	// Increment thumbs up count
	updateDoc := bson.M{
		"$inc": bson.M{"thumbs_up": 1},
		"$set": bson.M{"updated_at": time.Now().UTC()},
	}

	result, err := ideasCollection.UpdateOne(ctx, bson.M{"_id": ideaID}, updateDoc)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "DATABASE_ERROR",
				"message": "Failed to update thumbs up count",
				"details": err.Error(),
			},
		})
		return
	}

	if result.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": gin.H{
				"code":    "IDEA_NOT_FOUND",
				"message": "Idea not found",
			},
		})
		return
	}

	// Set rate limit
	setRateLimit(rateLimitKey, 5*time.Second)

	// Send notification to admin (async)
	go sendFeedbackNotification(idea.BoardID, ideaID, "thumbsup", clientIP)

	// Broadcast feedback animation to WebSocket clients
	utils.BroadcastFeedbackAnimation(idea.BoardID, ideaID, "thumbsup", "")

	// Return success response
	c.JSON(http.StatusOK, gin.H{
		"message":   "Thumbs up added successfully",
		"thumbsUp":  idea.ThumbsUp + 1,
		"timestamp": time.Now().UTC(),
	})
}

// AddEmojiReaction handles POST /api/ideas/:id/emoji (public endpoint)
func AddEmojiReaction(c *gin.Context) {
	// Get idea ID from URL parameter
	ideaID := c.Param("id")
	if ideaID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "INVALID_IDEA_ID",
				"message": "Idea ID is required",
			},
		})
		return
	}

	// Parse request body
	var req EmojiReactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "VALIDATION_ERROR",
				"message": "Invalid request data",
				"details": err.Error(),
			},
		})
		return
	}

	// Get client IP for rate limiting
	clientIP := c.ClientIP()

	// Rate limiting: check if this IP has made an emoji request in the last 3 seconds
	rateLimitKey := "emoji_" + ideaID + "_" + clientIP
	if isRateLimited(rateLimitKey, 3*time.Second) {
		c.JSON(http.StatusTooManyRequests, gin.H{
			"error": gin.H{
				"code":    "RATE_LIMITED",
				"message": "Please wait before adding another emoji reaction",
			},
		})
		return
	}

	// Basic emoji validation (prevent abuse)
	if !isValidEmoji(req.Emoji) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "INVALID_EMOJI",
				"message": "Invalid emoji provided",
			},
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Find the idea and verify it exists
	ideasCollection := models.GetCollection(models.IdeasCollection)
	var idea models.Idea
	err := ideasCollection.FindOne(ctx, bson.M{"_id": ideaID}).Decode(&idea)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{
				"error": gin.H{
					"code":    "IDEA_NOT_FOUND",
					"message": "Idea not found",
				},
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "DATABASE_ERROR",
				"message": "Failed to fetch idea",
				"details": err.Error(),
			},
		})
		return
	}

	// Update emoji reactions - increment existing or add new
	updateDoc := bson.M{
		"$set": bson.M{"updated_at": time.Now().UTC()},
	}

	// Check if emoji already exists in reactions
	emojiExists := false
	for i, reaction := range idea.EmojiReactions {
		if reaction.Emoji == req.Emoji {
			// Increment existing emoji count using array index
			updateDoc["$inc"] = bson.M{
				"emoji_reactions." + fmt.Sprintf("%d", i) + ".count": 1,
			}
			emojiExists = true
			break
		}
	}

	if !emojiExists {
		// Add new emoji reaction
		newReaction := models.EmojiReaction{
			Emoji: req.Emoji,
			Count: 1,
		}
		updateDoc["$push"] = bson.M{
			"emoji_reactions": newReaction,
		}
	}

	result, err := ideasCollection.UpdateOne(ctx, bson.M{"_id": ideaID}, updateDoc)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "DATABASE_ERROR",
				"message": "Failed to update emoji reaction",
				"details": err.Error(),
			},
		})
		return
	}

	if result.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": gin.H{
				"code":    "IDEA_NOT_FOUND",
				"message": "Idea not found",
			},
		})
		return
	}

	// Set rate limit
	setRateLimit(rateLimitKey, 3*time.Second)

	// Send notification to admin (async)
	go sendFeedbackNotification(idea.BoardID, ideaID, "emoji:"+req.Emoji, clientIP)

	// Broadcast feedback animation to WebSocket clients
	utils.BroadcastFeedbackAnimation(idea.BoardID, ideaID, "emoji", req.Emoji)

	// Return success response
	c.JSON(http.StatusOK, gin.H{
		"message":   "Emoji reaction added successfully",
		"emoji":     req.Emoji,
		"timestamp": time.Now().UTC(),
	})
}

// Simple in-memory rate limiting (for production, use Redis)
var rateLimitStore = make(map[string]time.Time)

func isRateLimited(key string, duration time.Duration) bool {
	if lastRequest, exists := rateLimitStore[key]; exists {
		if time.Since(lastRequest) < duration {
			return true
		}
	}
	return false
}

func setRateLimit(key string, duration time.Duration) {
	rateLimitStore[key] = time.Now()

	// Clean up old entries (simple cleanup)
	go func() {
		time.Sleep(duration * 2)
		delete(rateLimitStore, key)
	}()
}

// isValidEmoji performs basic emoji validation
func isValidEmoji(emoji string) bool {
	// Basic validation - check length and common emoji patterns
	if len(emoji) == 0 || len(emoji) > 10 {
		return false
	}

	// Allow common emoji characters (this is a simplified check)
	// In production, you'd want a more comprehensive emoji validation
	validEmojis := []string{
		"😀", "😃", "😄", "😁", "😆", "😅", "😂", "🤣", "😊", "😇",
		"🙂", "🙃", "😉", "😌", "😍", "🥰", "😘", "😗", "😙", "😚",
		"😋", "😛", "😝", "😜", "🤪", "🤨", "🧐", "🤓", "😎", "🤩",
		"🥳", "😏", "😒", "😞", "😔", "😟", "😕", "🙁", "☹️", "😣",
		"😖", "😫", "😩", "🥺", "😢", "😭", "😤", "😠", "😡", "🤬",
		"🤯", "😳", "🥵", "🥶", "😱", "😨", "😰", "😥", "😓", "🤗",
		"🤔", "🤭", "🤫", "🤥", "😶", "😐", "😑", "😬", "🙄", "😯",
		"😦", "😧", "😮", "😲", "🥱", "😴", "🤤", "😪", "😵", "🤐",
		"🥴", "🤢", "🤮", "🤧", "😷", "🤒", "🤕", "🤑", "🤠", "😈",
		"👍", "👎", "👌", "✌️", "🤞", "🤟", "🤘", "🤙", "👈", "👉",
		"👆", "🖕", "👇", "☝️", "👋", "🤚", "🖐️", "✋", "🖖", "👏",
		"🙌", "🤲", "🤝", "🙏", "✍️", "💪", "🦾", "🦿", "🦵", "🦶",
		"❤️", "🧡", "💛", "💚", "💙", "💜", "🖤", "🤍", "🤎", "💔",
		"❣️", "💕", "💞", "💓", "💗", "💖", "💘", "💝", "💟", "☮️",
		"✝️", "☪️", "🕉️", "☸️", "✡️", "🔯", "🕎", "☯️", "☦️", "🛐",
		"⭐", "🌟", "💫", "✨", "🌠", "🌙", "☀️", "🌤️", "⛅", "🌦️",
		"🌧️", "⛈️", "🌩️", "🌨️", "❄️", "☃️", "⛄", "🌬️", "💨", "🌪️",
		"🔥", "💥", "⚡", "🌈", "☔", "💧", "🌊", "🎉", "🎊", "🎈",
		"🎁", "🎀", "🏆", "🥇", "🥈", "🥉", "🏅", "🎖️", "🏵️", "🎗️",
	}

	for _, validEmoji := range validEmojis {
		if emoji == validEmoji {
			return true
		}
	}

	return false
}

// sendFeedbackNotification sends notifications to admin about feedback
func sendFeedbackNotification(boardID, ideaID, feedbackType, clientIP string) {
	// Use the notification service to send multi-channel notifications
	utils.SendFeedbackNotification(boardID, ideaID, feedbackType, clientIP)
}

// GetReleasedIdeasRequest represents query parameters for released ideas
type GetReleasedIdeasRequest struct {
	Search   string `form:"search"`
	SortBy   string `form:"sortBy"`  // name, created_at, thumbs_up, rice_score
	SortDir  string `form:"sortDir"` // asc, desc
	Page     int    `form:"page"`
	PageSize int    `form:"pageSize"`
}

// GetReleasedIdeas handles GET /api/boards/:id/release
func GetReleasedIdeas(c *gin.Context) {
	// Get board ID from URL parameter
	boardID := c.Param("id")
	if boardID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "INVALID_BOARD_ID",
				"message": "Board ID is required",
			},
		})
		return
	}

	// Parse query parameters
	var req GetReleasedIdeasRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "VALIDATION_ERROR",
				"message": "Invalid query parameters",
				"details": err.Error(),
			},
		})
		return
	}

	// Set defaults
	if req.SortBy == "" {
		req.SortBy = "created_at"
	}
	if req.SortDir == "" {
		req.SortDir = "desc"
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 50
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Check if this is a public request or admin request
	isPublic := c.GetHeader("X-Public-Access") == "true"

	if !isPublic {
		// For admin requests, verify board ownership
		userID, err := middleware.GetUserID(c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": gin.H{
					"code":    "INTERNAL_ERROR",
					"message": "Failed to get user ID",
				},
			})
			return
		}

		// Verify board exists and belongs to user
		boardsCollection := models.GetCollection(models.BoardsCollection)
		boardFilter := bson.M{
			"_id":     boardID,
			"user_id": userID,
		}

		var board models.Board
		err = boardsCollection.FindOne(ctx, boardFilter).Decode(&board)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				c.JSON(http.StatusNotFound, gin.H{
					"error": gin.H{
						"code":    "BOARD_NOT_FOUND",
						"message": "Board not found or you don't have permission to view ideas",
					},
				})
				return
			}

			c.JSON(http.StatusInternalServerError, gin.H{
				"error": gin.H{
					"code":    "DATABASE_ERROR",
					"message": "Failed to verify board",
					"details": err.Error(),
				},
			})
			return
		}
	} else {
		// For public requests, verify board exists by public link and is public
		boardsCollection := models.GetCollection(models.BoardsCollection)
		boardFilter := bson.M{"public_link": boardID, "is_public": true}

		var board models.Board
		err := boardsCollection.FindOne(ctx, boardFilter).Decode(&board)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				c.JSON(http.StatusNotFound, gin.H{
					"error": gin.H{
						"code":    "BOARD_NOT_FOUND",
						"message": "Board not found or is not publicly accessible. The board owner must make it public first.",
					},
				})
				return
			}

			c.JSON(http.StatusInternalServerError, gin.H{
				"error": gin.H{
					"code":    "DATABASE_ERROR",
					"message": "Failed to fetch board",
					"details": err.Error(),
				},
			})
			return
		}

		// Use the actual board ID for querying ideas
		boardID = board.ID
	}

	// Build filter for released ideas
	filter := bson.M{
		"board_id": boardID,
		"column":   string(models.ColumnRelease),
	}

	// Add search filter if provided
	if req.Search != "" {
		filter["$or"] = []bson.M{
			{"one_liner": bson.M{"$regex": req.Search, "$options": "i"}},
			{"description": bson.M{"$regex": req.Search, "$options": "i"}},
			{"value_statement": bson.M{"$regex": req.Search, "$options": "i"}},
		}
	}

	// Build sort options
	sortDir := 1
	if req.SortDir == "desc" {
		sortDir = -1
	}

	var sortField string
	switch req.SortBy {
	case "name":
		sortField = "one_liner"
	case "thumbs_up":
		sortField = "thumbs_up"
	case "rice_score":
		sortField = "rice_score.reach" // Sort by reach as primary RICE component
	default:
		sortField = "created_at"
	}

	opts := options.Find().
		SetSort(bson.D{{Key: sortField, Value: sortDir}}).
		SetSkip(int64((req.Page - 1) * req.PageSize)).
		SetLimit(int64(req.PageSize))

	// Query released ideas
	ideasCollection := models.GetCollection(models.IdeasCollection)
	cursor, err := ideasCollection.Find(ctx, filter, opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "DATABASE_ERROR",
				"message": "Failed to fetch released ideas",
				"details": err.Error(),
			},
		})
		return
	}
	defer cursor.Close(ctx)

	// Decode results
	var ideas []models.Idea
	if err := cursor.All(ctx, &ideas); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "DATABASE_ERROR",
				"message": "Failed to decode released ideas",
				"details": err.Error(),
			},
		})
		return
	}

	// Get total count for pagination
	totalCount, err := ideasCollection.CountDocuments(ctx, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "DATABASE_ERROR",
				"message": "Failed to count released ideas",
				"details": err.Error(),
			},
		})
		return
	}

	// Convert to response format
	var responses []interface{}
	for _, idea := range ideas {
		if isPublic {
			// Return public response format (filtered)
			responses = append(responses, PublicIdeaResponse{
				ID:             idea.ID,
				OneLiner:       idea.OneLiner,
				Description:    idea.Description,
				ValueStatement: idea.ValueStatement,
				Column:         idea.Column,
				Position:       idea.Position,
				InProgress:     idea.InProgress,
				ThumbsUp:       idea.ThumbsUp,
				EmojiReactions: idea.EmojiReactions,
				CreatedAt:      idea.CreatedAt,
				UpdatedAt:      idea.UpdatedAt,
			})
		} else {
			// Return full admin response format
			responses = append(responses, IdeaResponse{
				ID:             idea.ID,
				BoardID:        idea.BoardID,
				OneLiner:       idea.OneLiner,
				Description:    idea.Description,
				ValueStatement: idea.ValueStatement,
				RiceScore:      idea.RiceScore,
				Column:         idea.Column,
				Position:       idea.Position,
				InProgress:     idea.InProgress,
				Status:         idea.Status,
				ThumbsUp:       idea.ThumbsUp,
				EmojiReactions: idea.EmojiReactions,
				CreatedAt:      idea.CreatedAt,
				UpdatedAt:      idea.UpdatedAt,
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"ideas":      responses,
		"count":      len(responses),
		"totalCount": totalCount,
		"page":       req.Page,
		"pageSize":   req.PageSize,
		"totalPages": (int(totalCount) + req.PageSize - 1) / req.PageSize,
	})
}

// SearchBoardIdeasRequest represents the request parameters for searching ideas
type SearchBoardIdeasRequest struct {
	Query      string `form:"q"`
	SortBy     string `form:"sortBy"`     // "name", "rice", "status", "created"
	SortDir    string `form:"sortDir"`    // "asc", "desc"
	Column     string `form:"column"`     // filter by specific column
	Status     string `form:"status"`     // filter by status
	InProgress *bool  `form:"inProgress"` // filter by in-progress status
}

// SearchBoardIdeas handles GET /api/boards/:id/search
func SearchBoardIdeas(c *gin.Context) {
	// Get user ID from auth middleware
	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "INTERNAL_ERROR",
				"message": "Failed to get user ID",
			},
		})
		return
	}

	// Get board ID from URL parameter
	boardID := c.Param("id")
	if boardID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "INVALID_BOARD_ID",
				"message": "Board ID is required",
			},
		})
		return
	}

	// Parse query parameters
	var req SearchBoardIdeasRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "VALIDATION_ERROR",
				"message": "Invalid query parameters",
				"details": err.Error(),
			},
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Verify board exists and belongs to user
	boardsCollection := models.GetCollection(models.BoardsCollection)
	boardFilter := bson.M{
		"_id":     boardID,
		"user_id": userID,
	}

	var board models.Board
	err = boardsCollection.FindOne(ctx, boardFilter).Decode(&board)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{
				"error": gin.H{
					"code":    "BOARD_NOT_FOUND",
					"message": "Board not found or you don't have permission to search ideas",
				},
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "DATABASE_ERROR",
				"message": "Failed to verify board",
				"details": err.Error(),
			},
		})
		return
	}

	// Build aggregation pipeline
	pipeline := []bson.M{}

	// Match stage - filter by board ID
	matchStage := bson.M{
		"board_id": boardID,
	}

	// Add column filter if specified
	if req.Column != "" && models.IsValidColumn(req.Column) {
		matchStage["column"] = req.Column
	}

	// Add status filter if specified
	if req.Status != "" && models.IsValidStatus(req.Status) {
		matchStage["status"] = req.Status
	}

	// Add in-progress filter if specified
	if req.InProgress != nil {
		matchStage["in_progress"] = *req.InProgress
	}

	// Add text search if query is provided
	if req.Query != "" {
		// Use MongoDB regex search across multiple fields
		matchStage["$or"] = []bson.M{
			{"one_liner": bson.M{"$regex": req.Query, "$options": "i"}},
			{"description": bson.M{"$regex": req.Query, "$options": "i"}},
			{"value_statement": bson.M{"$regex": req.Query, "$options": "i"}},
		}
	}

	pipeline = append(pipeline, bson.M{"$match": matchStage})

	// Add calculated RICE score field for sorting
	pipeline = append(pipeline, bson.M{
		"$addFields": bson.M{
			"calculated_rice_score": bson.M{
				"$cond": bson.M{
					"if":   bson.M{"$eq": []interface{}{"$rice_score.effort", 0}},
					"then": 0,
					"else": bson.M{
						"$divide": []interface{}{
							bson.M{
								"$multiply": []interface{}{
									"$rice_score.reach",
									"$rice_score.impact",
									"$rice_score.confidence",
								},
							},
							"$rice_score.effort",
						},
					},
				},
			},
		},
	})

	// Add sorting
	sortStage := bson.M{}
	sortDirection := 1 // ascending by default
	if req.SortDir == "desc" {
		sortDirection = -1
	}

	switch req.SortBy {
	case "name":
		sortStage["one_liner"] = sortDirection
	case "rice":
		sortStage["calculated_rice_score"] = sortDirection
	case "status":
		// Sort by in_progress first, then by status
		sortStage["in_progress"] = -1 // in-progress items first
		sortStage["status"] = sortDirection
	case "created":
		sortStage["created_at"] = sortDirection
	default:
		// Default sort: column, then position
		sortStage["column"] = 1
		sortStage["position"] = 1
	}

	pipeline = append(pipeline, bson.M{"$sort": sortStage})

	// Execute aggregation
	ideasCollection := models.GetCollection(models.IdeasCollection)
	cursor, err := ideasCollection.Aggregate(ctx, pipeline)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "DATABASE_ERROR",
				"message": "Failed to search ideas",
				"details": err.Error(),
			},
		})
		return
	}
	defer cursor.Close(ctx)

	// Decode results
	var ideas []models.Idea
	if err := cursor.All(ctx, &ideas); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "DATABASE_ERROR",
				"message": "Failed to decode search results",
				"details": err.Error(),
			},
		})
		return
	}

	// Convert to response format
	var responses []IdeaResponse
	for _, idea := range ideas {
		responses = append(responses, IdeaResponse{
			ID:             idea.ID,
			BoardID:        idea.BoardID,
			OneLiner:       idea.OneLiner,
			Description:    idea.Description,
			ValueStatement: idea.ValueStatement,
			RiceScore:      idea.RiceScore,
			Column:         idea.Column,
			Position:       idea.Position,
			InProgress:     idea.InProgress,
			Status:         idea.Status,
			ThumbsUp:       idea.ThumbsUp,
			EmojiReactions: idea.EmojiReactions,
			CreatedAt:      idea.CreatedAt,
			UpdatedAt:      idea.UpdatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"ideas": responses,
		"count": len(responses),
		"query": req.Query,
		"filters": gin.H{
			"column":     req.Column,
			"status":     req.Status,
			"inProgress": req.InProgress,
		},
		"sort": gin.H{
			"by":        req.SortBy,
			"direction": req.SortDir,
		},
	})
}
