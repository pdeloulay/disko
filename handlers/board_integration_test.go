//go:build integration
// +build integration

package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"disko-backend/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestMain(m *testing.M) {
	// Set up test environment
	os.Setenv("MONGODB_URI", "mongodb://localhost:27017")
	os.Setenv("MONGODB_DATABASE", "disko_board_test")
	os.Setenv("CLERK_SECRET_KEY", "test_secret_key")

	// Initialize test database
	if err := models.ConnectDatabase(); err != nil {
		panic("Failed to connect to test database: " + err.Error())
	}

	// Run tests
	code := m.Run()

	// Clean up
	cleanupTestData()
	models.DisconnectDatabase()

	os.Exit(code)
}

func cleanupTestData() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Clean up test data
	models.GetCollection(models.BoardsCollection).Drop(ctx)
	models.GetCollection(models.IdeasCollection).Drop(ctx)
}

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Add test middleware to simulate authenticated user
	router.Use(func(c *gin.Context) {
		c.Set("userID", "test_user_123")
		c.Set("sessionID", "test_session_123")
		c.Next()
	})

	return router
}

func TestCreateBoard(t *testing.T) {
	router := setupTestRouter()
	router.POST("/api/boards", CreateBoard)

	tests := []struct {
		name           string
		requestBody    CreateBoardRequest
		expectedStatus int
		expectedError  string
	}{
		{
			name: "Valid board creation",
			requestBody: CreateBoardRequest{
				Name:        "Test Board",
				Description: "Test Description",
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "Board creation with custom columns",
			requestBody: CreateBoardRequest{
				Name:           "Custom Board",
				Description:    "Custom Description",
				VisibleColumns: []string{"parking", "now", "next"},
				VisibleFields:  []string{"oneLiner", "description"},
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "Invalid board - empty name",
			requestBody: CreateBoardRequest{
				Name:        "",
				Description: "Test Description",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "VALIDATION_ERROR",
		},
		{
			name: "Invalid board - name too long",
			requestBody: CreateBoardRequest{
				Name:        string(make([]byte, 101)), // 101 characters
				Description: "Test Description",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "VALIDATION_ERROR",
		},
		{
			name: "Invalid column type",
			requestBody: CreateBoardRequest{
				Name:           "Test Board",
				VisibleColumns: []string{"invalid_column"},
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "INVALID_COLUMN",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up before each test
			cleanupTestData()

			// Prepare request
			jsonBody, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest("POST", "/api/boards", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")

			// Execute request
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Assert status code
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusCreated {
				// Parse response
				var response BoardResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)

				// Validate response
				assert.Equal(t, tt.requestBody.Name, response.Name)
				assert.Equal(t, tt.requestBody.Description, response.Description)
				assert.NotEmpty(t, response.ID)
				assert.NotEmpty(t, response.PublicLink)
				assert.Equal(t, "test_user_123", response.AdminID)
				assert.NotEmpty(t, response.VisibleColumns)
				assert.NotEmpty(t, response.VisibleFields)
			} else if tt.expectedError != "" {
				// Parse error response
				var errorResponse map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
				require.NoError(t, err)

				errorObj := errorResponse["error"].(map[string]interface{})
				assert.Equal(t, tt.expectedError, errorObj["code"])
			}
		})
	}
}

func TestGetBoards(t *testing.T) {
	router := setupTestRouter()
	router.GET("/api/boards", GetBoards)

	// Clean up and create test data
	cleanupTestData()
	createTestBoard(t, "Test Board 1", "test_user_123")
	createTestBoard(t, "Test Board 2", "test_user_123")
	createTestBoard(t, "Other User Board", "other_user_456")

	// Execute request
	req, _ := http.NewRequest("GET", "/api/boards", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	boards := response["boards"].([]interface{})
	count := response["count"].(float64)

	// Should only return boards for the authenticated user
	assert.Equal(t, 2, int(count))
	assert.Len(t, boards, 2)

	// Verify board data
	board1 := boards[0].(map[string]interface{})
	assert.Equal(t, "test_user_123", board1["adminId"])
}

func createTestBoard(t *testing.T, name, adminID string) string {
	boardID := "b" + uuid.New().String()[:8]
	board := models.Board{
		ID:             boardID,
		Name:           name,
		Description:    "Test Description",
		PublicLink:     "test-public-link-" + name,
		AdminID:        adminID,
		VisibleColumns: models.GetDefaultVisibleColumns(),
		VisibleFields:  models.GetDefaultVisibleFields(),
		CreatedAt:      time.Now().UTC(),
		UpdatedAt:      time.Now().UTC(),
	}

	collection := models.GetCollection(models.BoardsCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := collection.InsertOne(ctx, board)
	require.NoError(t, err)

	return boardID
}
func TestUpdateBoard(t *testing.T) {
	router := setupTestRouter()
	router.PUT("/api/boards/:id", UpdateBoard)

	// Clean up and create test board
	cleanupTestData()
	boardID := createTestBoard(t, "Original Board", "test_user_123")

	tests := []struct {
		name           string
		boardID        string
		requestBody    UpdateBoardRequest
		expectedStatus int
		expectedError  string
	}{
		{
			name:    "Valid board update",
			boardID: boardID,
			requestBody: UpdateBoardRequest{
				Name:        "Updated Board",
				Description: "Updated Description",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:    "Update visible columns",
			boardID: boardID,
			requestBody: UpdateBoardRequest{
				VisibleColumns: []string{"parking", "now"},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid board ID",
			boardID:        "invalid_id",
			requestBody:    UpdateBoardRequest{Name: "Updated"},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "INVALID_BOARD_ID",
		},
		{
			name:           "Board not found",
			boardID:        "b" + uuid.New().String()[:8],
			requestBody:    UpdateBoardRequest{Name: "Updated"},
			expectedStatus: http.StatusNotFound,
			expectedError:  "BOARD_NOT_FOUND",
		},
		{
			name:    "Invalid column type",
			boardID: boardID,
			requestBody: UpdateBoardRequest{
				VisibleColumns: []string{"invalid_column"},
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "INVALID_COLUMN",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Prepare request
			jsonBody, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest("PUT", "/api/boards/"+tt.boardID, bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")

			// Execute request
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Assert status code
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				// Parse response
				var response BoardResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)

				// Validate updated fields
				if tt.requestBody.Name != "" {
					assert.Equal(t, tt.requestBody.Name, response.Name)
				}
				if tt.requestBody.Description != "" {
					assert.Equal(t, tt.requestBody.Description, response.Description)
				}
				if len(tt.requestBody.VisibleColumns) > 0 {
					assert.Equal(t, tt.requestBody.VisibleColumns, response.VisibleColumns)
				}
			} else if tt.expectedError != "" {
				// Parse error response
				var errorResponse map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
				require.NoError(t, err)

				errorObj := errorResponse["error"].(map[string]interface{})
				assert.Equal(t, tt.expectedError, errorObj["code"])
			}
		})
	}
}

func TestDeleteBoard(t *testing.T) {
	router := setupTestRouter()
	router.DELETE("/api/boards/:id", DeleteBoard)

	tests := []struct {
		name           string
		setupBoard     bool
		boardID        string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "Valid board deletion",
			setupBoard:     true,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid board ID",
			setupBoard:     false,
			boardID:        "invalid_id",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "INVALID_BOARD_ID",
		},
		{
			name:           "Board not found",
			setupBoard:     false,
			boardID:        "b" + uuid.New().String()[:8],
			expectedStatus: http.StatusNotFound,
			expectedError:  "BOARD_NOT_FOUND",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up and setup test data
			cleanupTestData()

			var boardID string
			if tt.setupBoard {
				boardID = createTestBoard(t, "Board to Delete", "test_user_123")

				// Create some test ideas for cascade deletion test
				createTestIdea(t, boardID, "Test Idea 1")
				createTestIdea(t, boardID, "Test Idea 2")
			} else {
				boardID = tt.boardID
			}

			// Execute request
			req, _ := http.NewRequest("DELETE", "/api/boards/"+boardID, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Assert status code
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				// Verify board is deleted
				collection := models.GetCollection(models.BoardsCollection)
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()

				var board models.Board
				err := collection.FindOne(ctx, bson.M{"_id": boardID}).Decode(&board)
				assert.Error(t, err) // Should not find the board

				// Verify ideas are also deleted (cascade)
				ideasCollection := models.GetCollection(models.IdeasCollection)
				cursor, err := ideasCollection.Find(ctx, bson.M{"board_id": boardID})
				require.NoError(t, err)
				defer cursor.Close(ctx)

				var ideas []bson.M
				err = cursor.All(ctx, &ideas)
				require.NoError(t, err)
				assert.Empty(t, ideas) // Should have no ideas left
			} else if tt.expectedError != "" {
				// Parse error response
				var errorResponse map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
				require.NoError(t, err)

				errorObj := errorResponse["error"].(map[string]interface{})
				assert.Equal(t, tt.expectedError, errorObj["code"])
			}
		})
	}
}

func createTestIdea(t *testing.T, boardID string, title string) {
	idea := bson.M{
		"board_id":        boardID,
		"one_liner":       title,
		"description":     "Test description",
		"value_statement": "Test value",
		"rice_score": bson.M{
			"reach":      50,
			"impact":     60,
			"confidence": 2,
			"effort":     40,
		},
		"column":          "parking",
		"position":        1,
		"in_progress":     false,
		"status":          "draft",
		"thumbs_up":       0,
		"emoji_reactions": []bson.M{},
		"created_at":      time.Now().UTC(),
		"updated_at":      time.Now().UTC(),
	}

	collection := models.GetCollection(models.IdeasCollection)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := collection.InsertOne(ctx, idea)
	require.NoError(t, err)
}
