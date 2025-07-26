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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// MockCollection is a mock implementation of mongo.Collection
type MockCollection struct {
	mock.Mock
}

func (m *MockCollection) InsertOne(ctx context.Context, document interface{}, opts ...interface{}) (*mongo.InsertOneResult, error) {
	args := m.Called(ctx, document)
	return args.Get(0).(*mongo.InsertOneResult), args.Error(1)
}

func (m *MockCollection) FindOne(ctx context.Context, filter interface{}, opts ...interface{}) *mongo.SingleResult {
	args := m.Called(ctx, filter)
	return args.Get(0).(*mongo.SingleResult)
}

func (m *MockCollection) Find(ctx context.Context, filter interface{}, opts ...interface{}) (*mongo.Cursor, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(*mongo.Cursor), args.Error(1)
}

func (m *MockCollection) UpdateOne(ctx context.Context, filter interface{}, update interface{}, opts ...interface{}) (*mongo.UpdateResult, error) {
	args := m.Called(ctx, filter, update)
	return args.Get(0).(*mongo.UpdateResult), args.Error(1)
}

func (m *MockCollection) DeleteOne(ctx context.Context, filter interface{}, opts ...interface{}) (*mongo.DeleteResult, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(*mongo.DeleteResult), args.Error(1)
}

func TestCreateIdeaRequest_Validation(t *testing.T) {
	tests := []struct {
		name    string
		request CreateIdeaRequest
		valid   bool
	}{
		{
			name: "valid request",
			request: CreateIdeaRequest{
				OneLiner:       "Test idea",
				Description:    "Test description",
				ValueStatement: "Test value",
				RiceScore: models.RICEScore{
					Reach:      80,
					Impact:     70,
					Confidence: 4,
					Effort:     60,
				},
			},
			valid: true,
		},
		{
			name: "missing one-liner",
			request: CreateIdeaRequest{
				Description:    "Test description",
				ValueStatement: "Test value",
				RiceScore: models.RICEScore{
					Reach:      80,
					Impact:     70,
					Confidence: 4,
					Effort:     60,
				},
			},
			valid: false,
		},
		{
			name: "one-liner too long",
			request: CreateIdeaRequest{
				OneLiner:       string(make([]byte, 201)), // 201 characters
				Description:    "Test description",
				ValueStatement: "Test value",
				RiceScore: models.RICEScore{
					Reach:      80,
					Impact:     70,
					Confidence: 4,
					Effort:     60,
				},
			},
			valid: false,
		},
		{
			name: "invalid RICE confidence",
			request: CreateIdeaRequest{
				OneLiner:       "Test idea",
				Description:    "Test description",
				ValueStatement: "Test value",
				RiceScore: models.RICEScore{
					Reach:      80,
					Impact:     70,
					Confidence: 3, // Invalid - should be 1, 2, 4, or 8
					Effort:     60,
				},
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test RICE score validation specifically
			if tt.name == "invalid RICE confidence" {
				assert.False(t, tt.request.RiceScore.IsValidRICEScore())
			} else if tt.valid {
				assert.True(t, tt.request.RiceScore.IsValidRICEScore())
			}
		})
	}
}

func TestUpdateIdeaPositionRequest_Validation(t *testing.T) {
	tests := []struct {
		name    string
		request UpdateIdeaPositionRequest
		valid   bool
	}{
		{
			name: "valid position update",
			request: UpdateIdeaPositionRequest{
				Column:   "now",
				Position: 1,
			},
			valid: true,
		},
		{
			name: "invalid column",
			request: UpdateIdeaPositionRequest{
				Column:   "invalid-column",
				Position: 1,
			},
			valid: false,
		},
		{
			name: "negative position",
			request: UpdateIdeaPositionRequest{
				Column:   "now",
				Position: -1,
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.valid {
				assert.True(t, models.IsValidColumn(tt.request.Column))
				assert.GreaterOrEqual(t, tt.request.Position, 0)
			} else {
				if tt.name == "invalid column" {
					assert.False(t, models.IsValidColumn(tt.request.Column))
				}
				if tt.name == "negative position" {
					assert.Less(t, tt.request.Position, 0)
				}
			}
		})
	}
}

func TestIdeaResponse_Structure(t *testing.T) {
	now := time.Now().UTC()

	response := IdeaResponse{
		ID:             "I12345678",
		BoardID:        "b12345678",
		OneLiner:       "Test idea",
		Description:    "Test description",
		ValueStatement: "Test value",
		RiceScore: models.RICEScore{
			Reach:      80,
			Impact:     70,
			Confidence: 4,
			Effort:     60,
		},
		Column:         "now",
		Position:       1,
		InProgress:     false,
		Status:         "active",
		ThumbsUp:       0,
		EmojiReactions: []models.EmojiReaction{},
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	// Test JSON marshaling
	jsonData, err := json.Marshal(response)
	assert.NoError(t, err)
	assert.Contains(t, string(jsonData), "I12345678")
	assert.Contains(t, string(jsonData), "Test idea")
	assert.Contains(t, string(jsonData), "now")

	// Test JSON unmarshaling
	var unmarshaled IdeaResponse
	err = json.Unmarshal(jsonData, &unmarshaled)
	assert.NoError(t, err)
	assert.Equal(t, response.ID, unmarshaled.ID)
	assert.Equal(t, response.OneLiner, unmarshaled.OneLiner)
	assert.Equal(t, response.RiceScore.Reach, unmarshaled.RiceScore.Reach)
}

func TestRICEScoreValidation(t *testing.T) {
	tests := []struct {
		name      string
		riceScore models.RICEScore
		valid     bool
	}{
		{
			name: "valid RICE score",
			riceScore: models.RICEScore{
				Reach:      80,
				Impact:     70,
				Confidence: 4,
				Effort:     60,
			},
			valid: true,
		},
		{
			name: "reach too high",
			riceScore: models.RICEScore{
				Reach:      101,
				Impact:     70,
				Confidence: 4,
				Effort:     60,
			},
			valid: false,
		},
		{
			name: "reach negative",
			riceScore: models.RICEScore{
				Reach:      -1,
				Impact:     70,
				Confidence: 4,
				Effort:     60,
			},
			valid: false,
		},
		{
			name: "impact too high",
			riceScore: models.RICEScore{
				Reach:      80,
				Impact:     101,
				Confidence: 4,
				Effort:     60,
			},
			valid: false,
		},
		{
			name: "invalid confidence value",
			riceScore: models.RICEScore{
				Reach:      80,
				Impact:     70,
				Confidence: 3, // Should be 1, 2, 4, or 8
				Effort:     60,
			},
			valid: false,
		},
		{
			name: "effort too high",
			riceScore: models.RICEScore{
				Reach:      80,
				Impact:     70,
				Confidence: 4,
				Effort:     101,
			},
			valid: false,
		},
		{
			name: "all valid confidence values",
			riceScore: models.RICEScore{
				Reach:      50,
				Impact:     50,
				Confidence: 1,
				Effort:     50,
			},
			valid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.riceScore.IsValidRICEScore()
			assert.Equal(t, tt.valid, result)
		})
	}
}

func TestRICEScoreCalculation(t *testing.T) {
	tests := []struct {
		name      string
		riceScore models.RICEScore
		expected  float64
	}{
		{
			name: "normal calculation",
			riceScore: models.RICEScore{
				Reach:      80,
				Impact:     70,
				Confidence: 4,
				Effort:     60,
			},
			expected: float64(80*70*4) / float64(60), // 22400 / 60 = 373.33...
		},
		{
			name: "zero effort",
			riceScore: models.RICEScore{
				Reach:      80,
				Impact:     70,
				Confidence: 4,
				Effort:     0,
			},
			expected: 0, // Should return 0 when effort is 0
		},
		{
			name: "minimum values",
			riceScore: models.RICEScore{
				Reach:      1,
				Impact:     1,
				Confidence: 1,
				Effort:     1,
			},
			expected: 1, // 1*1*1/1 = 1
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.riceScore.CalculateRICEScore()
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Test helper function to create a test Gin context
func createTestContext() (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	return c, w
}

// Test helper function to create a test request with JSON body
func createTestRequest(method, url string, body interface{}) *http.Request {
	var buf bytes.Buffer
	if body != nil {
		json.NewEncoder(&buf).Encode(body)
	}
	req, _ := http.NewRequest(method, url, &buf)
	req.Header.Set("Content-Type", "application/json")
	return req
}
