package handlers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"disko-backend/middleware"
	"disko-backend/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// CreateBoardRequest represents the request payload for creating a board
type CreateBoardRequest struct {
	Name           string   `json:"name" binding:"required,min=1,max=100"`
	Description    string   `json:"description,omitempty" binding:"max=500"`
	VisibleColumns []string `json:"visibleColumns,omitempty"`
	VisibleFields  []string `json:"visibleFields,omitempty"`
}

// UpdateBoardRequest represents the request payload for updating a board
type UpdateBoardRequest struct {
	Name           string   `json:"name,omitempty" binding:"omitempty,min=1,max=100"`
	Description    string   `json:"description,omitempty" binding:"max=500"`
	VisibleColumns []string `json:"visibleColumns,omitempty"`
	VisibleFields  []string `json:"visibleFields,omitempty"`
}

// BoardResponse represents the response format for board operations
type BoardResponse struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	Description    string    `json:"description,omitempty"`
	PublicLink     string    `json:"publicLink"`
	AdminID        string    `json:"adminId"`
	VisibleColumns []string  `json:"visibleColumns"`
	VisibleFields  []string  `json:"visibleFields"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

// CreateBoard handles POST /api/boards
func CreateBoard(c *gin.Context) {
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

	// Parse request body
	var req CreateBoardRequest
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

	// Set defaults if not provided
	visibleColumns := req.VisibleColumns
	if len(visibleColumns) == 0 {
		visibleColumns = models.GetDefaultVisibleColumns()
	}

	visibleFields := req.VisibleFields
	if len(visibleFields) == 0 {
		visibleFields = models.GetDefaultVisibleFields()
	}

	// Validate visible columns
	for _, column := range visibleColumns {
		if !models.IsValidColumn(column) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": gin.H{
					"code":    "INVALID_COLUMN",
					"message": "Invalid column type: " + column,
				},
			})
			return
		}
	}

	// Generate unique public link using UUID
	publicLink := uuid.New().String()

	// Generate unique board ID using short UUID with "b" prefix
	boardID := "b" + uuid.New().String()[:8]

	// Create board document
	now := time.Now().UTC()
	board := models.Board{
		ID:             boardID,
		Name:           req.Name,
		Description:    req.Description,
		PublicLink:     publicLink,
		AdminID:        userID,
		VisibleColumns: visibleColumns,
		VisibleFields:  visibleFields,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	// Insert into MongoDB
	collection := models.GetCollection(models.BoardsCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err = collection.InsertOne(ctx, board)
	if err != nil {
		// Check if it's a duplicate public link error (very unlikely with UUID)
		if mongo.IsDuplicateKeyError(err) {
			c.JSON(http.StatusConflict, gin.H{
				"error": gin.H{
					"code":    "DUPLICATE_PUBLIC_LINK",
					"message": "Public link already exists, please try again",
				},
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "DATABASE_ERROR",
				"message": "Failed to create board",
				"details": err.Error(),
			},
		})
		return
	}

	// Return created board
	response := BoardResponse{
		ID:             board.ID,
		Name:           board.Name,
		Description:    board.Description,
		PublicLink:     board.PublicLink,
		AdminID:        board.AdminID,
		VisibleColumns: board.VisibleColumns,
		VisibleFields:  board.VisibleFields,
		CreatedAt:      board.CreatedAt,
		UpdatedAt:      board.UpdatedAt,
	}

	c.JSON(http.StatusCreated, response)
}

// GetBoards handles GET /api/boards
func GetBoards(c *gin.Context) {
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

	// Query boards for the authenticated user
	collection := models.GetCollection(models.BoardsCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"admin_id": userID}
	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "DATABASE_ERROR",
				"message": "Failed to fetch boards",
				"details": err.Error(),
			},
		})
		return
	}
	defer cursor.Close(ctx)

	// Decode results
	var boards []models.Board
	if err := cursor.All(ctx, &boards); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "DATABASE_ERROR",
				"message": "Failed to decode boards",
				"details": err.Error(),
			},
		})
		return
	}

	// Convert to response format
	var responses []BoardResponse
	for _, board := range boards {
		responses = append(responses, BoardResponse{
			ID:             board.ID,
			Name:           board.Name,
			Description:    board.Description,
			PublicLink:     board.PublicLink,
			AdminID:        board.AdminID,
			VisibleColumns: board.VisibleColumns,
			VisibleFields:  board.VisibleFields,
			CreatedAt:      board.CreatedAt,
			UpdatedAt:      board.UpdatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"boards": responses,
		"count":  len(responses),
	})
}

// UpdateBoard handles PUT /api/boards/:id
func UpdateBoard(c *gin.Context) {
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
	var req UpdateBoardRequest
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

	// Build update document
	updateDoc := bson.M{
		"updated_at": time.Now().UTC(),
	}

	if req.Name != "" {
		updateDoc["name"] = req.Name
	}

	if req.Description != "" {
		updateDoc["description"] = req.Description
	}

	if len(req.VisibleColumns) > 0 {
		// Validate visible columns
		for _, column := range req.VisibleColumns {
			if !models.IsValidColumn(column) {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": gin.H{
						"code":    "INVALID_COLUMN",
						"message": "Invalid column type: " + column,
					},
				})
				return
			}
		}
		updateDoc["visible_columns"] = req.VisibleColumns
	}

	if len(req.VisibleFields) > 0 {
		updateDoc["visible_fields"] = req.VisibleFields
	}

	// Update board in MongoDB
	collection := models.GetCollection(models.BoardsCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"_id":      boardID,
		"admin_id": userID, // Ensure user can only update their own boards
	}

	result, err := collection.UpdateOne(ctx, filter, bson.M{"$set": updateDoc})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "DATABASE_ERROR",
				"message": "Failed to update board",
				"details": err.Error(),
			},
		})
		return
	}

	if result.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": gin.H{
				"code":    "BOARD_NOT_FOUND",
				"message": "Board not found or you don't have permission to update it",
			},
		})
		return
	}

	// Fetch and return updated board
	var updatedBoard models.Board
	err = collection.FindOne(ctx, filter).Decode(&updatedBoard)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "DATABASE_ERROR",
				"message": "Failed to fetch updated board",
				"details": err.Error(),
			},
		})
		return
	}

	// Return updated board
	response := BoardResponse{
		ID:             updatedBoard.ID,
		Name:           updatedBoard.Name,
		Description:    updatedBoard.Description,
		PublicLink:     updatedBoard.PublicLink,
		AdminID:        updatedBoard.AdminID,
		VisibleColumns: updatedBoard.VisibleColumns,
		VisibleFields:  updatedBoard.VisibleFields,
		CreatedAt:      updatedBoard.CreatedAt,
		UpdatedAt:      updatedBoard.UpdatedAt,
	}

	c.JSON(http.StatusOK, response)
}

// DeleteBoard handles DELETE /api/boards/:id
func DeleteBoard(c *gin.Context) {
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

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Start a transaction for cascade deletion
	session, err := models.DB.Client.StartSession()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "DATABASE_ERROR",
				"message": "Failed to start database transaction",
				"details": err.Error(),
			},
		})
		return
	}
	defer session.EndSession(ctx)

	// Execute transaction
	err = mongo.WithSession(ctx, session, func(sc context.Context) error {
		// First, verify the board exists and belongs to the user
		boardsCollection := models.GetCollection(models.BoardsCollection)
		boardFilter := bson.M{
			"_id":      boardID,
			"admin_id": userID,
		}

		var board models.Board
		err := boardsCollection.FindOne(sc, boardFilter).Decode(&board)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return &BoardNotFoundError{}
			}
			return err
		}

		// Delete all ideas associated with this board
		ideasCollection := models.GetCollection(models.IdeasCollection)
		ideasFilter := bson.M{"board_id": boardID}

		ideasResult, err := ideasCollection.DeleteMany(sc, ideasFilter)
		if err != nil {
			return err
		}

		// Delete the board
		boardResult, err := boardsCollection.DeleteOne(sc, boardFilter)
		if err != nil {
			return err
		}

		if boardResult.DeletedCount == 0 {
			return &BoardNotFoundError{}
		}

		// Log the cascade deletion for debugging
		c.Header("X-Ideas-Deleted", fmt.Sprintf("%d", ideasResult.DeletedCount))

		return nil
	})

	if err != nil {
		if _, ok := err.(*BoardNotFoundError); ok {
			c.JSON(http.StatusNotFound, gin.H{
				"error": gin.H{
					"code":    "BOARD_NOT_FOUND",
					"message": "Board not found or you don't have permission to delete it",
				},
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "DATABASE_ERROR",
				"message": "Failed to delete board",
				"details": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Board and all associated ideas deleted successfully",
	})
}

// BoardNotFoundError represents a board not found error
type BoardNotFoundError struct{}

func (e *BoardNotFoundError) Error() string {
	return "board not found"
}
