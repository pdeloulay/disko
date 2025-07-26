//go:build integration
// +build integration

package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"disko-backend/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestCreateIdeaIntegration(t *testing.T) {
	// Set up test board
	testUserID := "test_user_" + uuid.New().String()[:8]
	testBoardID := "b" + uuid.New().String()[:8]

	// Create test board
	board := models.Board{
		ID:             testBoardID,
		Name:           "Test Board",
		Description:    "Test Description",
		PublicLink:     uuid.New().String(),
		AdminID:        testUserID,
		VisibleColumns: models.GetDefaultVisibleColumns(),
		VisibleFields:  models.GetDefaultVisibleFields(),
		CreatedAt:      time.Now().UTC(),
		UpdatedAt:      time.Now().UTC(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	boardsCollection := models.GetCollection(models.BoardsCollection)
	_, err := boardsCollection.InsertOne(ctx, board)
	require.NoError(t, err)

	// Clean up after test
	defer func() {
		boardsCollection.DeleteOne(ctx, bson.M{"_id": testBoardID})
		ideasCollection := models.GetCollection(models.IdeasCollection)
		ideasCollection.DeleteMany(ctx, bson.M{"board_id": testBoardID})
	}()

	// Test create idea
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Mock auth middleware for testing
	router.Use(func(c *gin.Context) {
		c.Set("userID", testUserID)
		c.Next()
	})

	router.POST("/api/boards/:id/ideas", CreateIdea)

	// Create test request
	ideaRequest := CreateIdeaRequest{
		OneLiner:       "Test Idea",
		Description:    "This is a test idea description",
		ValueStatement: "This provides test value",
		RiceScore: models.RICEScore{
			Reach:      80,
			Impact:     70,
			Confidence: 4,
			Effort:     60,
		},
		Column:   "parking",
		Position: 1,
	}

	jsonData, err := json.Marshal(ideaRequest)
	require.NoError(t, err)

	req, err := http.NewRequest("POST", "/api/boards/"+testBoardID+"/ideas", bytes.NewBuffer(jsonData))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Verify response
	assert.Equal(t, http.StatusCreated, w.Code)

	var response IdeaResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.NotEmpty(t, response.ID)
	assert.True(t, len(response.ID) > 1 && response.ID[0] == 'I') // Should start with "I"
	assert.Equal(t, testBoardID, response.BoardID)
	assert.Equal(t, ideaRequest.OneLiner, response.OneLiner)
	assert.Equal(t, ideaRequest.Description, response.Description)
	assert.Equal(t, ideaRequest.ValueStatement, response.ValueStatement)
	assert.Equal(t, ideaRequest.RiceScore, response.RiceScore)
	assert.Equal(t, "parking", response.Column) // Should default to parking
	assert.Equal(t, 1, response.Position)
	assert.False(t, response.InProgress)
	assert.Equal(t, "active", response.Status)
	assert.Equal(t, 0, response.ThumbsUp)
	assert.Empty(t, response.EmojiReactions)
}

func TestGetBoardIdeasIntegration(t *testing.T) {
	// Set up test board and ideas
	testUserID := "test_user_" + uuid.New().String()[:8]
	testBoardID := "b" + uuid.New().String()[:8]

	// Create test board
	board := models.Board{
		ID:             testBoardID,
		Name:           "Test Board",
		Description:    "Test Description",
		PublicLink:     uuid.New().String(),
		AdminID:        testUserID,
		VisibleColumns: models.GetDefaultVisibleColumns(),
		VisibleFields:  models.GetDefaultVisibleFields(),
		CreatedAt:      time.Now().UTC(),
		UpdatedAt:      time.Now().UTC(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	boardsCollection := models.GetCollection(models.BoardsCollection)
	_, err := boardsCollection.InsertOne(ctx, board)
	require.NoError(t, err)

	// Create test ideas
	ideas := []models.Idea{
		{
			ID:             "I" + uuid.New().String()[:8],
			BoardID:        testBoardID,
			OneLiner:       "First Idea",
			Description:    "First description",
			ValueStatement: "First value",
			RiceScore: models.RICEScore{
				Reach:      80,
				Impact:     70,
				Confidence: 4,
				Effort:     60,
			},
			Column:         "parking",
			Position:       1,
			InProgress:     false,
			Status:         "active",
			ThumbsUp:       0,
			EmojiReactions: []models.EmojiReaction{},
			CreatedAt:      time.Now().UTC(),
			UpdatedAt:      time.Now().UTC(),
		},
		{
			ID:             "I" + uuid.New().String()[:8],
			BoardID:        testBoardID,
			OneLiner:       "Second Idea",
			Description:    "Second description",
			ValueStatement: "Second value",
			RiceScore: models.RICEScore{
				Reach:      90,
				Impact:     80,
				Confidence: 2,
				Effort:     50,
			},
			Column:         "now",
			Position:       1,
			InProgress:     true,
			Status:         "active",
			ThumbsUp:       5,
			EmojiReactions: []models.EmojiReaction{},
			CreatedAt:      time.Now().UTC(),
			UpdatedAt:      time.Now().UTC(),
		},
	}

	ideasCollection := models.GetCollection(models.IdeasCollection)
	for _, idea := range ideas {
		_, err := ideasCollection.InsertOne(ctx, idea)
		require.NoError(t, err)
	}

	// Clean up after test
	defer func() {
		boardsCollection.DeleteOne(ctx, bson.M{"_id": testBoardID})
		ideasCollection.DeleteMany(ctx, bson.M{"board_id": testBoardID})
	}()

	// Test get board ideas
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Mock auth middleware for testing
	router.Use(func(c *gin.Context) {
		c.Set("userID", testUserID)
		c.Next()
	})

	router.GET("/api/boards/:id/ideas", GetBoardIdeas)

	req, err := http.NewRequest("GET", "/api/boards/"+testBoardID+"/ideas", nil)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Verify response
	assert.Equal(t, http.StatusOK, w.Code)

	var response struct {
		Ideas []IdeaResponse `json:"ideas"`
		Count int            `json:"count"`
	}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, 2, response.Count)
	assert.Len(t, response.Ideas, 2)

	// Verify ideas are sorted by column and position
	assert.Equal(t, "First Idea", response.Ideas[0].OneLiner)
	assert.Equal(t, "parking", response.Ideas[0].Column)
	assert.Equal(t, "Second Idea", response.Ideas[1].OneLiner)
	assert.Equal(t, "now", response.Ideas[1].Column)
	assert.True(t, response.Ideas[1].InProgress)
}

func TestUpdateIdeaIntegration(t *testing.T) {
	// Set up test board and idea
	testUserID := "test_user_" + uuid.New().String()[:8]
	testBoardID := "b" + uuid.New().String()[:8]
	testIdeaID := "I" + uuid.New().String()[:8]

	// Create test board
	board := models.Board{
		ID:             testBoardID,
		Name:           "Test Board",
		Description:    "Test Description",
		PublicLink:     uuid.New().String(),
		AdminID:        testUserID,
		VisibleColumns: models.GetDefaultVisibleColumns(),
		VisibleFields:  models.GetDefaultVisibleFields(),
		CreatedAt:      time.Now().UTC(),
		UpdatedAt:      time.Now().UTC(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	boardsCollection := models.GetCollection(models.BoardsCollection)
	_, err := boardsCollection.InsertOne(ctx, board)
	require.NoError(t, err)

	// Create test idea
	idea := models.Idea{
		ID:             testIdeaID,
		BoardID:        testBoardID,
		OneLiner:       "Original Idea",
		Description:    "Original description",
		ValueStatement: "Original value",
		RiceScore: models.RICEScore{
			Reach:      80,
			Impact:     70,
			Confidence: 4,
			Effort:     60,
		},
		Column:         "parking",
		Position:       1,
		InProgress:     false,
		Status:         "active",
		ThumbsUp:       0,
		EmojiReactions: []models.EmojiReaction{},
		CreatedAt:      time.Now().UTC(),
		UpdatedAt:      time.Now().UTC(),
	}

	ideasCollection := models.GetCollection(models.IdeasCollection)
	_, err = ideasCollection.InsertOne(ctx, idea)
	require.NoError(t, err)

	// Clean up after test
	defer func() {
		boardsCollection.DeleteOne(ctx, bson.M{"_id": testBoardID})
		ideasCollection.DeleteMany(ctx, bson.M{"board_id": testBoardID})
	}()

	// Test update idea
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Mock auth middleware for testing
	router.Use(func(c *gin.Context) {
		c.Set("userID", testUserID)
		c.Next()
	})

	router.PUT("/api/ideas/:id", UpdateIdea)

	// Create update request
	inProgress := true
	updateRequest := UpdateIdeaRequest{
		OneLiner:   "Updated Idea",
		InProgress: &inProgress,
		Status:     "done", // This should move it to release column
	}

	jsonData, err := json.Marshal(updateRequest)
	require.NoError(t, err)

	req, err := http.NewRequest("PUT", "/api/ideas/"+testIdeaID, bytes.NewBuffer(jsonData))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Verify response
	assert.Equal(t, http.StatusOK, w.Code)

	var response IdeaResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, testIdeaID, response.ID)
	assert.Equal(t, "Updated Idea", response.OneLiner)
	assert.True(t, response.InProgress)
	assert.Equal(t, "done", response.Status)
	assert.Equal(t, "release", response.Column) // Should be moved to release when status is done
}

func TestRICEScoreValidationInAPI(t *testing.T) {
	// Set up test board
	testUserID := "test_user_" + uuid.New().String()[:8]
	testBoardID := "b" + uuid.New().String()[:8]

	// Create test board
	board := models.Board{
		ID:             testBoardID,
		Name:           "Test Board",
		Description:    "Test Description",
		PublicLink:     uuid.New().String(),
		AdminID:        testUserID,
		VisibleColumns: models.GetDefaultVisibleColumns(),
		VisibleFields:  models.GetDefaultVisibleFields(),
		CreatedAt:      time.Now().UTC(),
		UpdatedAt:      time.Now().UTC(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	boardsCollection := models.GetCollection(models.BoardsCollection)
	_, err := boardsCollection.InsertOne(ctx, board)
	require.NoError(t, err)

	// Clean up after test
	defer func() {
		boardsCollection.DeleteOne(ctx, bson.M{"_id": testBoardID})
		ideasCollection := models.GetCollection(models.IdeasCollection)
		ideasCollection.DeleteMany(ctx, bson.M{"board_id": testBoardID})
	}()

	// Test invalid RICE score
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Mock auth middleware for testing
	router.Use(func(c *gin.Context) {
		c.Set("userID", testUserID)
		c.Next()
	})

	router.POST("/api/boards/:id/ideas", CreateIdea)

	// Create request with invalid RICE score
	ideaRequest := CreateIdeaRequest{
		OneLiner:       "Test Idea",
		Description:    "This is a test idea description",
		ValueStatement: "This provides test value",
		RiceScore: models.RICEScore{
			Reach:      101, // Invalid - should be 0-100
			Impact:     70,
			Confidence: 3, // Invalid - should be 1, 2, 4, or 8
			Effort:     60,
		},
	}

	jsonData, err := json.Marshal(ideaRequest)
	require.NoError(t, err)

	req, err := http.NewRequest("POST", "/api/boards/"+testBoardID+"/ideas", bytes.NewBuffer(jsonData))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Verify error response
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var errorResponse map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &errorResponse)
	require.NoError(t, err)

	assert.Contains(t, errorResponse, "error")
	errorObj := errorResponse["error"].(map[string]interface{})
	assert.Equal(t, "INVALID_RICE_SCORE", errorObj["code"])
	assert.Contains(t, errorObj["message"], "Invalid RICE score values")
}
