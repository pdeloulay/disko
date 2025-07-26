package handlers

import (
	"context"
	"net/http"
	"time"

	"disko-backend/middleware"
	"disko-backend/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
		"_id":      boardID,
		"admin_id": userID,
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
	ideaID := "I" + uuid.New().String()[:8]

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

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Verify board exists and belongs to user
	boardsCollection := models.GetCollection(models.BoardsCollection)
	boardFilter := bson.M{
		"_id":      boardID,
		"admin_id": userID,
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

	// Query ideas for the board
	ideasCollection := models.GetCollection(models.IdeasCollection)
	ideasFilter := bson.M{"board_id": boardID}

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
		"_id":      existingIdea.BoardID,
		"admin_id": userID,
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

		// If status is "done", move to release column
		if req.Status == string(models.StatusDone) {
			updateDoc["column"] = string(models.ColumnRelease)
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
		"_id":      existingIdea.BoardID,
		"admin_id": userID,
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
		"_id":      existingIdea.BoardID,
		"admin_id": userID,
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
		"_id":      existingIdea.BoardID,
		"admin_id": userID,
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

	c.JSON(http.StatusOK, response)
}
