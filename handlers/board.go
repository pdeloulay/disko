package handlers

import (
	"context"
	"log"
	"net/http"
	"time"

	"disko-backend/middleware"
	"disko-backend/models"
	"disko-backend/utils"

	"github.com/gin-gonic/gin"
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
	IsPublic       *bool    `json:"isPublic,omitempty"`
}

// BoardResponse represents the response format for board operations
type BoardResponse struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	Description    string    `json:"description,omitempty"`
	PublicLink     string    `json:"publicLink"`
	IsPublic       bool      `json:"isPublic"`
	UserID         string    `json:"userId"`
	IsAdmin        bool      `json:"isAdmin"`
	VisibleColumns []string  `json:"visibleColumns"`
	VisibleFields  []string  `json:"visibleFields"`
	IdeasCount     int       `json:"ideasCount"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

// CreateBoard handles POST /api/boards
func CreateBoard(c *gin.Context) {
	startTime := time.Now()
	userAgent := c.GetHeader("User-Agent")
	referer := c.GetHeader("Referer")

	// Get user ID from auth middleware
	userID, err := middleware.GetUserID(c)
	if err != nil {
		log.Printf("[Handler] CreateBoard failed - GetUserID error: %v, IP: %s, UserAgent: %s", err, c.ClientIP(), userAgent)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "INTERNAL_ERROR",
				"message": "Failed to get user ID",
			},
		})
		return
	}

	log.Printf("[Handler] CreateBoard started - UserID: %s, IP: %s, UserAgent: %s, Referer: %s",
		userID, c.ClientIP(), userAgent, referer)

	// Parse request body
	parseStartTime := time.Now()
	var req CreateBoardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		parseDuration := time.Since(parseStartTime)
		log.Printf("[Handler] CreateBoard failed - JSON binding error: %v, UserID: %s, Duration: %v, IP: %s",
			err, userID, parseDuration, c.ClientIP())
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "VALIDATION_ERROR",
				"message": "Invalid request data",
				"details": err.Error(),
			},
		})
		return
	}
	parseDuration := time.Since(parseStartTime)

	log.Printf("[Handler] CreateBoard - Request parsed successfully - Name: %s, Description: %s, VisibleColumns: %v, VisibleFields: %v, UserID: %s, Parse duration: %v",
		req.Name, req.Description, req.VisibleColumns, req.VisibleFields, userID, parseDuration)

	// Set defaults if not provided
	configStartTime := time.Now()
	visibleColumns := req.VisibleColumns
	if len(visibleColumns) == 0 {
		visibleColumns = models.GetDefaultVisibleColumns()
		log.Printf("[Handler] CreateBoard - Using default visible columns: %v, UserID: %s", visibleColumns, userID)
	}

	visibleFields := req.VisibleFields
	if len(visibleFields) == 0 {
		visibleFields = models.GetDefaultVisibleFields()
		log.Printf("[Handler] CreateBoard - Using default visible fields: %v, UserID: %s", visibleFields, userID)
	}
	configDuration := time.Since(configStartTime)
	log.Printf("[Handler] CreateBoard - Configuration completed - Duration: %v, UserID: %s", configDuration, userID)

	// Validate visible columns
	validationStartTime := time.Now()
	for _, column := range visibleColumns {
		if !models.IsValidColumn(column) {
			validationDuration := time.Since(validationStartTime)
			log.Printf("[Handler] CreateBoard failed - Invalid column: %s, UserID: %s, Duration: %v, IP: %s",
				column, userID, validationDuration, c.ClientIP())
			c.JSON(http.StatusBadRequest, gin.H{
				"error": gin.H{
					"code":    "INVALID_COLUMN",
					"message": "Invalid column type: " + column,
				},
			})
			return
		}
	}
	validationDuration := time.Since(validationStartTime)
	log.Printf("[Handler] CreateBoard - Column validation successful - Duration: %v, UserID: %s", validationDuration, userID)

	// Generate unique public link using short Google UUID
	generateStartTime := time.Now()
	publicLink := utils.GenerateShortUUID()
	boardID := utils.GenerateBoardID()
	generateDuration := time.Since(generateStartTime)

	log.Printf("[Handler] CreateBoard - Generated IDs - BoardID: %s, PublicLink: %s, Duration: %v, UserID: %s",
		boardID, publicLink, generateDuration, userID)

	// Create board document
	now := time.Now().UTC()
	board := models.Board{
		ID:             boardID,
		Name:           req.Name,
		Description:    req.Description,
		PublicLink:     publicLink,
		IsPublic:       false, // Boards are private by default
		UserID:         userID,
		VisibleColumns: visibleColumns,
		VisibleFields:  visibleFields,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	// Insert into MongoDB
	collection := models.GetCollection(models.BoardsCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	log.Printf("[Handler] CreateBoard - Collection insertion - Database: disko, Collection: boards, UserID: %s, BoardID: %s",
		userID, boardID)

	dbStartTime := time.Now()
	_, err = collection.InsertOne(ctx, board)
	dbDuration := time.Since(dbStartTime)

	if err != nil {
		// Check if it's a duplicate public link error (very unlikely with UUID)
		if mongo.IsDuplicateKeyError(err) {
			log.Printf("[Handler] CreateBoard failed - Duplicate key error: %v, UserID: %s, Duration: %v, IP: %s",
				err, userID, dbDuration, c.ClientIP())
			c.JSON(http.StatusConflict, gin.H{
				"error": gin.H{
					"code":    "DUPLICATE_PUBLIC_LINK",
					"message": "Public link already exists, please try again",
				},
			})
			return
		}

		log.Printf("[Handler] CreateBoard failed - Database insert error: %v, UserID: %s, Duration: %v, IP: %s",
			err, userID, dbDuration, c.ClientIP())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "DATABASE_ERROR",
				"message": "Failed to create board",
				"details": err.Error(),
			},
		})
		return
	}

	log.Printf("[Handler] CreateBoard - Collection insertion successful - Board added to collection: ID=%s, Name=%s, UserID: %s, Duration: %v",
		boardID, board.Name, userID, dbDuration)

	// Create default idea for the new board
	defaultIdeaStartTime := time.Now()
	defaultIdea := models.Idea{
		ID:             utils.GenerateIdeaID(),
		BoardID:        boardID,
		OneLiner:       "Welcome to your new board! 🎉",
		Description:    "This is your first idea. Click to edit and start building your roadmap.",
		ValueStatement: "Get started by adding your first real idea to this board.",
		RiceScore: models.RICEScore{
			Reach:      50,
			Impact:     50,
			Confidence: 4,
			Effort:     50,
		},
		Column:         "now",
		Position:       1,
		InProgress:     false,
		Status:         string(models.StatusActive),
		ThumbsUp:       0,
		EmojiReactions: []models.EmojiReaction{},
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	// Insert default idea
	ideasCollection := models.GetCollection(models.IdeasCollection)
	_, err = ideasCollection.InsertOne(ctx, defaultIdea)
	if err != nil {
		log.Printf("[Handler] CreateBoard - Failed to create default idea: %v, BoardID: %s, UserID: %s", err, boardID, userID)
		// Don't fail the board creation if default idea fails
	} else {
		defaultIdeaDuration := time.Since(defaultIdeaStartTime)
		log.Printf("[Handler] CreateBoard - Default idea created successfully - IdeaID: %s, BoardID: %s, Duration: %v, UserID: %s",
			defaultIdea.ID, boardID, defaultIdeaDuration, userID)
	}

	// Create response
	responseStartTime := time.Now()
	response := BoardResponse{
		ID:             board.ID,
		Name:           board.Name,
		Description:    board.Description,
		PublicLink:     board.PublicLink,
		IsPublic:       board.IsPublic,
		UserID:         board.UserID,
		VisibleColumns: board.VisibleColumns,
		VisibleFields:  board.VisibleFields,
		CreatedAt:      board.CreatedAt,
		UpdatedAt:      board.UpdatedAt,
	}
	responseDuration := time.Since(responseStartTime)

	totalDuration := time.Since(startTime)
	log.Printf("[Handler] CreateBoard completed successfully - BoardID: %s, Name: %s, Total duration: %v, Response duration: %v, UserID: %s, IP: %s",
		board.ID, board.Name, totalDuration, responseDuration, userID, c.ClientIP())

	c.JSON(http.StatusCreated, response)
}

// GetBoards handles GET /api/boards
func GetBoards(c *gin.Context) {
	startTime := time.Now()
	userAgent := c.GetHeader("User-Agent")
	referer := c.GetHeader("Referer")

	// Get user ID from auth middleware
	userID, err := middleware.GetUserID(c)
	if err != nil {
		log.Printf("[Handler] GetBoards failed - GetUserID error: %v, IP: %s, UserAgent: %s", err, c.ClientIP(), userAgent)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "INTERNAL_ERROR",
				"message": "Failed to get user ID",
			},
		})
		return
	}

	log.Printf("[Handler] GetBoards started - UserID: %s, IP: %s, UserAgent: %s, Referer: %s",
		userID, c.ClientIP(), userAgent, referer)

	// Query boards for the authenticated user
	collection := models.GetCollection(models.BoardsCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"user_id": userID}
	log.Printf("[Handler] GetBoards - Executing database query - Filter: %v, UserID: %s", filter, userID)

	// Log collection details
	log.Printf("[Handler] GetBoards - Collection lookup - Database: disko, Collection: boards, UserID: %s", userID)

	dbStartTime := time.Now()
	cursor, err := collection.Find(ctx, filter)
	dbDuration := time.Since(dbStartTime)

	if err != nil {
		log.Printf("[Handler] GetBoards failed - Database query error: %v, UserID: %s, Duration: %v, IP: %s",
			err, userID, dbDuration, c.ClientIP())
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

	log.Printf("[Handler] GetBoards - Database query successful - Duration: %v, UserID: %s", dbDuration, userID)

	// Decode results
	decodeStartTime := time.Now()
	var boards []models.Board
	if err := cursor.All(ctx, &boards); err != nil {
		decodeDuration := time.Since(decodeStartTime)
		log.Printf("[Handler] GetBoards failed - Decode error: %v, UserID: %s, Duration: %v, IP: %s",
			err, userID, decodeDuration, c.ClientIP())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "DATABASE_ERROR",
				"message": "Failed to decode boards",
				"details": err.Error(),
			},
		})
		return
	}
	decodeDuration := time.Since(decodeStartTime)

	log.Printf("[Handler] GetBoards - Collection lookup results - Boards found: %d, UserID: %s, Decode duration: %v",
		len(boards), userID, decodeDuration)

	// Log detailed board information
	if len(boards) > 0 {
		log.Printf("[Handler] GetBoards - Board collection details for UserID %s:", userID)
		for i, board := range boards {
			log.Printf("[Handler] GetBoards - Board %d/%d: ID=%s, Name=%s, PublicLink=%s, CreatedAt=%s, UpdatedAt=%s",
				i+1, len(boards), board.ID, board.Name, board.PublicLink,
				board.CreatedAt.Format("2006-01-02 15:04:05"),
				board.UpdatedAt.Format("2006-01-02 15:04:05"))
		}
	} else {
		log.Printf("[Handler] GetBoards - No boards found in collection for UserID: %s", userID)
	}

	// Convert to response format and count ideas for each board
	responseStartTime := time.Now()
	var responses []BoardResponse
	for i, board := range boards {
		// Count ideas for this board
		ideasCollection := models.GetCollection(models.IdeasCollection)
		ideasFilter := bson.M{"board_id": board.ID}
		ideasCount, err := ideasCollection.CountDocuments(ctx, ideasFilter)
		if err != nil {
			log.Printf("[Handler] GetBoards - Failed to count ideas for board %s: %v", board.ID, err)
			ideasCount = 0
		}

		responses = append(responses, BoardResponse{
			ID:             board.ID,
			Name:           board.Name,
			Description:    board.Description,
			PublicLink:     board.PublicLink,
			IsPublic:       board.IsPublic,
			UserID:         board.UserID,
			VisibleColumns: board.VisibleColumns,
			VisibleFields:  board.VisibleFields,
			IdeasCount:     int(ideasCount),
			CreatedAt:      board.CreatedAt,
			UpdatedAt:      board.UpdatedAt,
		})
		log.Printf("[Handler] GetBoards - Board %d: ID=%s, Name=%s, PublicLink=%s, IdeasCount=%d",
			i+1, board.ID, board.Name, board.PublicLink, ideasCount)
	}
	responseDuration := time.Since(responseStartTime)

	totalDuration := time.Since(startTime)
	log.Printf("[Handler] GetBoards completed successfully - Collection lookup summary: Total boards: %d, UserID: %s, Total duration: %v, Response duration: %v, IP: %s",
		len(responses), userID, totalDuration, responseDuration, c.ClientIP())

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
		// Validate visible fields
		for _, field := range req.VisibleFields {
			if !models.IsValidField(field) {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": gin.H{
						"code":    "INVALID_FIELD",
						"message": "Invalid field type: " + field,
					},
				})
				return
			}
		}
		updateDoc["visible_fields"] = req.VisibleFields
	}

	// Handle isPublic field
	if req.IsPublic != nil {
		updateDoc["is_public"] = *req.IsPublic

		// If setting to public, generate new public link for enhanced security
		if *req.IsPublic {
			newPublicLink := utils.GenerateShortUUID()
			updateDoc["public_link"] = newPublicLink
			log.Printf("[Handler] UpdateBoard - Generating new public link for board: %s, NewLink: %s", boardID, newPublicLink)
		}
	}

	// Update board in MongoDB
	collection := models.GetCollection(models.BoardsCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"_id":     boardID,
		"user_id": userID, // Ensure user can only update their own boards
	}

	log.Printf("[Handler] UpdateBoard - Collection update - Database: disko, Collection: boards, BoardID: %s, UserID: %s, UpdateDoc: %v",
		boardID, userID, updateDoc)

	updateStartTime := time.Now()
	result, err := collection.UpdateOne(ctx, filter, bson.M{"$set": updateDoc})
	updateDuration := time.Since(updateStartTime)

	if err != nil {
		log.Printf("[Handler] UpdateBoard failed - Collection update error: %v, BoardID: %s, UserID: %s, Duration: %v",
			err, boardID, userID, updateDuration)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "DATABASE_ERROR",
				"message": "Failed to update board",
				"details": err.Error(),
			},
		})
		return
	}

	log.Printf("[Handler] UpdateBoard - Collection update successful - Matched: %d, Modified: %d, BoardID: %s, UserID: %s, Duration: %v",
		result.MatchedCount, result.ModifiedCount, boardID, userID, updateDuration)

	if result.MatchedCount == 0 {
		log.Printf("[Handler] UpdateBoard failed - Board not found in collection - BoardID: %s, UserID: %s", boardID, userID)
		c.JSON(http.StatusNotFound, gin.H{
			"error": gin.H{
				"code":    "BOARD_NOT_FOUND",
				"message": "Board not found or you don't have permission to update it",
			},
		})
		return
	}

	// Fetch and return updated board
	log.Printf("[Handler] UpdateBoard - Fetching updated board from collection - BoardID: %s, UserID: %s", boardID, userID)

	fetchStartTime := time.Now()
	var updatedBoard models.Board
	err = collection.FindOne(ctx, filter).Decode(&updatedBoard)
	fetchDuration := time.Since(fetchStartTime)

	if err != nil {
		log.Printf("[Handler] UpdateBoard failed - Fetch updated board error: %v, BoardID: %s, UserID: %s, Duration: %v",
			err, boardID, userID, fetchDuration)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "DATABASE_ERROR",
				"message": "Failed to fetch updated board",
				"details": err.Error(),
			},
		})
		return
	}

	log.Printf("[Handler] UpdateBoard - Updated board fetched from collection - BoardID: %s, Name: %s, UserID: %s, Duration: %v",
		updatedBoard.ID, updatedBoard.Name, userID, fetchDuration)

	// Return updated board
	response := BoardResponse{
		ID:             updatedBoard.ID,
		Name:           updatedBoard.Name,
		Description:    updatedBoard.Description,
		PublicLink:     updatedBoard.PublicLink,
		UserID:         updatedBoard.UserID,
		VisibleColumns: updatedBoard.VisibleColumns,
		VisibleFields:  updatedBoard.VisibleFields,
		CreatedAt:      updatedBoard.CreatedAt,
		UpdatedAt:      updatedBoard.UpdatedAt,
	}

	c.JSON(http.StatusOK, response)
}

// DeleteBoard handles DELETE /api/boards/:id
func DeleteBoard(c *gin.Context) {
	startTime := time.Now()
	userAgent := c.GetHeader("User-Agent")
	referer := c.GetHeader("Referer")

	// Get user ID from auth middleware
	userID, err := middleware.GetUserID(c)
	if err != nil {
		log.Printf("[Handler] DeleteBoard failed - GetUserID error: %v, IP: %s, UserAgent: %s", err, c.ClientIP(), userAgent)
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
		log.Printf("[Handler] DeleteBoard failed - Invalid board ID: empty, UserID: %s, IP: %s", userID, c.ClientIP())
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "INVALID_BOARD_ID",
				"message": "Board ID is required",
			},
		})
		return
	}

	log.Printf("[Handler] DeleteBoard started - BoardID: %s, UserID: %s, IP: %s, UserAgent: %s, Referer: %s",
		boardID, userID, c.ClientIP(), userAgent, referer)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Start a transaction for cascade deletion
	sessionStartTime := time.Now()
	session, err := models.DB.Client.StartSession()
	if err != nil {
		sessionDuration := time.Since(sessionStartTime)
		log.Printf("[Handler] DeleteBoard failed - Session start error: %v, BoardID: %s, UserID: %s, Duration: %v, IP: %s",
			err, boardID, userID, sessionDuration, c.ClientIP())
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
	sessionDuration := time.Since(sessionStartTime)
	log.Printf("[Handler] DeleteBoard - Database session started - Duration: %v, BoardID: %s, UserID: %s",
		sessionDuration, boardID, userID)

	// Execute transaction
	transactionStartTime := time.Now()
	err = mongo.WithSession(ctx, session, func(sc context.Context) error {
		// First, verify the board exists and belongs to the user
		boardsCollection := models.GetCollection(models.BoardsCollection)
		boardFilter := bson.M{
			"_id":     boardID,
			"user_id": userID,
		}

		log.Printf("[Handler] DeleteBoard - Verifying board ownership - Filter: %v, BoardID: %s, UserID: %s",
			boardFilter, boardID, userID)

		var board models.Board
		err := boardsCollection.FindOne(sc, boardFilter).Decode(&board)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				log.Printf("[Handler] DeleteBoard failed - Board not found or access denied - BoardID: %s, UserID: %s",
					boardID, userID)
				return &BoardNotFoundError{}
			}
			log.Printf("[Handler] DeleteBoard failed - Board verification error: %v, BoardID: %s, UserID: %s",
				err, boardID, userID)
			return err
		}

		log.Printf("[Handler] DeleteBoard - Board verified - Name: %s, PublicLink: %s, BoardID: %s, UserID: %s",
			board.Name, board.PublicLink, boardID, userID)

		// Delete all ideas associated with this board
		ideasCollection := models.GetCollection(models.IdeasCollection)
		ideasFilter := bson.M{"board_id": boardID}

		log.Printf("[Handler] DeleteBoard - Collection deletion - Ideas collection: Database: disko, Collection: ideas, BoardID: %s, UserID: %s",
			boardID, userID)

		ideasResult, err := ideasCollection.DeleteMany(sc, ideasFilter)
		if err != nil {
			log.Printf("[Handler] DeleteBoard failed - Ideas deletion error: %v, BoardID: %s, UserID: %s",
				err, boardID, userID)
			return err
		}

		log.Printf("[Handler] DeleteBoard - Ideas collection deletion successful - Ideas deleted: %d, BoardID: %s, UserID: %s",
			ideasResult.DeletedCount, boardID, userID)

		// Delete the board itself
		log.Printf("[Handler] DeleteBoard - Collection deletion - Boards collection: Database: disko, Collection: boards, BoardID: %s, UserID: %s",
			boardID, userID)

		boardResult, err := boardsCollection.DeleteOne(sc, boardFilter)
		if err != nil {
			log.Printf("[Handler] DeleteBoard failed - Board deletion error: %v, BoardID: %s, UserID: %s",
				err, boardID, userID)
			return err
		}

		log.Printf("[Handler] DeleteBoard - Boards collection deletion successful - Board deleted: %d, BoardID: %s, UserID: %s",
			boardResult.DeletedCount, boardID, userID)

		return nil
	})
	transactionDuration := time.Since(transactionStartTime)

	if err != nil {
		log.Printf("[Handler] DeleteBoard failed - Transaction error: %v, BoardID: %s, UserID: %s, Duration: %v, IP: %s",
			err, boardID, userID, transactionDuration, c.ClientIP())

		if _, ok := err.(*BoardNotFoundError); ok {
			c.JSON(http.StatusNotFound, gin.H{
				"error": gin.H{
					"code":    "BOARD_NOT_FOUND",
					"message": "Board not found or access denied",
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

	totalDuration := time.Since(startTime)
	log.Printf("[Handler] DeleteBoard completed successfully - BoardID: %s, UserID: %s, Transaction duration: %v, Total duration: %v, IP: %s",
		boardID, userID, transactionDuration, totalDuration, c.ClientIP())

	c.JSON(http.StatusOK, gin.H{
		"message": "Board deleted successfully",
		"boardID": boardID,
	})
}

// PublicBoardResponse represents the response format for public board access
type PublicBoardResponse struct {
	ID             string    `json:"id"`
	Name           string    `json:"name"`
	Description    string    `json:"description,omitempty"`
	VisibleColumns []string  `json:"visibleColumns"`
	VisibleFields  []string  `json:"visibleFields"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

// GetBoard handles GET /api/boards/:id (for authenticated users)
func GetBoard(c *gin.Context) {

	startTime := time.Now()
	boardID := c.Param("id")
	log.Printf("[Handler] GetBoard - BoardID: %s", boardID)

	userAgent := c.GetHeader("User-Agent")
	referer := c.GetHeader("Referer")

	// Get user ID from auth middleware
	userID, err := middleware.GetUserID(c)
	if err != nil {
		log.Printf("[Handler] GetBoard failed - GetUserID error: %v, BoardID: %s, IP: %s, UserAgent: %s", err, boardID, c.ClientIP(), userAgent)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "INTERNAL_ERROR",
				"message": "Failed to get user ID",
			},
		})
		return
	}

	log.Printf("[Handler] GetBoard started - BoardID: %s, UserID: %s, IP: %s, UserAgent: %s, Referer: %s",
		boardID, userID, c.ClientIP(), userAgent, referer)

	// Get database connection
	if models.DB == nil {
		log.Printf("[Handler] GetBoard failed - Database connection failed, BoardID: %s, UserID: %s", boardID, userID)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "DATABASE_ERROR",
				"message": "Database connection failed",
			},
		})
		return
	}

	// Find the board
	collection := models.GetCollection(models.BoardsCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"_id": boardID, "user_id": userID}
	log.Printf("[Handler] GetBoard - Database query: Filter: %+v, BoardID: %s, UserID: %s", filter, boardID, userID)
	log.Printf("[Handler] GetBoard - Database connection status: %t", models.DB != nil)
	log.Printf("[Handler] GetBoard - Collection name: %s", models.BoardsCollection)

	var board models.Board
	if err := collection.FindOne(ctx, filter).Decode(&board); err != nil {
		if err == mongo.ErrNoDocuments {
			log.Printf("[Handler] GetBoard failed - Board not found or user does not own it: BoardID: %s, UserID: %s, Error: %v", boardID, userID, err)
			c.JSON(http.StatusNotFound, gin.H{
				"error": gin.H{
					"code":    "BOARD_NOT_FOUND",
					"message": "Board not found or you don't have permission to access it",
				},
			})
		} else {
			log.Printf("[Handler] GetBoard failed - Database error: BoardID: %s, UserID: %s, Error: %v", boardID, userID, err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": gin.H{
					"code":    "DATABASE_ERROR",
					"message": "Failed to retrieve board",
				},
			})
		}
		return
	}

	// Convert to response format
	response := BoardResponse{
		ID:             board.ID,
		Name:           board.Name,
		Description:    board.Description,
		PublicLink:     board.PublicLink,
		IsPublic:       board.IsPublic,
		UserID:         board.UserID,
		IsAdmin:        board.UserID == userID, // User is admin if they own the board
		VisibleColumns: board.VisibleColumns,
		VisibleFields:  board.VisibleFields,
		CreatedAt:      board.CreatedAt,
		UpdatedAt:      board.UpdatedAt,
	}

	duration := time.Since(startTime)
	log.Printf("[Handler] GetBoard success - BoardID: %s, UserID: %s, Duration: %v, IP: %s",
		boardID, userID, duration, c.ClientIP())
	log.Printf("[Handler] GetBoard - Board details: ID=%s, Name=%s, PublicLink=%s, IsPublic=%t, UserID=%s",
		board.ID, board.Name, board.PublicLink, board.IsPublic, board.UserID)

	c.JSON(http.StatusOK, response)
}

// GetPublicBoard handles GET /api/boards/:id/public
func GetPublicBoard(c *gin.Context) {
	startTime := time.Now()
	userAgent := c.GetHeader("User-Agent")
	referer := c.GetHeader("Referer")

	// Get public link from URL parameter
	publicLink := c.Param("id")
	if publicLink == "" {
		log.Printf("[Handler] GetPublicBoard failed - Invalid public link: empty, IP: %s, UserAgent: %s", c.ClientIP(), userAgent)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "INVALID_PUBLIC_LINK",
				"message": "Public link is required",
			},
		})
		return
	}

	log.Printf("[Handler] GetPublicBoard started - PublicLink: %s, IP: %s, UserAgent: %s, Referer: %s",
		publicLink, c.ClientIP(), userAgent, referer)

	// Query board by public link
	collection := models.GetCollection(models.BoardsCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"public_link": publicLink, "is_public": true}
	log.Printf("[Handler] GetPublicBoard - Collection lookup - Database: disko, Collection: boards, PublicLink: %s, Filter: %v",
		publicLink, filter)

	dbStartTime := time.Now()
	var board models.Board
	err := collection.FindOne(ctx, filter).Decode(&board)
	dbDuration := time.Since(dbStartTime)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			log.Printf("[Handler] GetPublicBoard failed - Board not found or not public - PublicLink: %s, Duration: %v, IP: %s",
				publicLink, dbDuration, c.ClientIP())
			c.JSON(http.StatusNotFound, gin.H{
				"error": gin.H{
					"code":    "BOARD_NOT_FOUND",
					"message": "Board not found or is not publicly accessible. The board owner must make it public first.",
				},
			})
			return
		}

		log.Printf("[Handler] GetPublicBoard failed - Collection lookup error: %v, PublicLink: %s, Duration: %v, IP: %s",
			err, publicLink, dbDuration, c.ClientIP())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "DATABASE_ERROR",
				"message": "Failed to fetch board",
				"details": err.Error(),
			},
		})
		return
	}

	log.Printf("[Handler] GetPublicBoard - Collection lookup successful - Board found: ID=%s, Name=%s, PublicLink=%s, Duration: %v",
		board.ID, board.Name, board.PublicLink, dbDuration)

	// Return public board data (without admin-only information)
	responseStartTime := time.Now()
	response := PublicBoardResponse{
		ID:             board.ID,
		Name:           board.Name,
		Description:    board.Description,
		VisibleColumns: board.VisibleColumns,
		VisibleFields:  board.VisibleFields,
		CreatedAt:      board.CreatedAt,
		UpdatedAt:      board.UpdatedAt,
	}
	responseDuration := time.Since(responseStartTime)

	totalDuration := time.Since(startTime)
	log.Printf("[Handler] GetPublicBoard completed successfully - Collection lookup summary: BoardID: %s, Name: %s, Total duration: %v, Response duration: %v, IP: %s",
		board.ID, board.Name, totalDuration, responseDuration, c.ClientIP())

	c.JSON(http.StatusOK, response)
}

// GetPublicReleasedIdeas handles GET /api/boards/:id/release/public
func GetPublicReleasedIdeas(c *gin.Context) {
	boardID := c.Param("id")
	log.Printf("[API] GetReleasedIdeas (public) called - BoardID: %s, IP: %s, UserAgent: %s", boardID, c.ClientIP(), c.GetHeader("User-Agent"))
	c.Header("X-Public-Access", "true")
	GetReleasedIdeas(c)
}

// BoardNotFoundError represents a board not found error
type BoardNotFoundError struct{}

func (e *BoardNotFoundError) Error() string {
	return "board not found"
}
