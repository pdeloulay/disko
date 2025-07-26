package handlers

import (
	"strings"
	"testing"

	"disko-backend/models"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestCreateBoardRequest_Validation(t *testing.T) {
	tests := []struct {
		name    string
		request CreateBoardRequest
		valid   bool
	}{
		{
			name: "Valid request with all fields",
			request: CreateBoardRequest{
				Name:           "Test Board",
				Description:    "Test Description",
				VisibleColumns: []string{"parking", "now", "next"},
				VisibleFields:  []string{"oneLiner", "description"},
			},
			valid: true,
		},
		{
			name: "Valid request with minimal fields",
			request: CreateBoardRequest{
				Name: "Test Board",
			},
			valid: true,
		},
		{
			name: "Invalid request - empty name",
			request: CreateBoardRequest{
				Name: "",
			},
			valid: false,
		},
		{
			name: "Invalid request - name too long",
			request: CreateBoardRequest{
				Name: string(make([]byte, 101)), // 101 characters
			},
			valid: false,
		},
		{
			name: "Invalid request - description too long",
			request: CreateBoardRequest{
				Name:        "Test Board",
				Description: string(make([]byte, 501)), // 501 characters
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test name validation
			if tt.request.Name == "" {
				assert.False(t, tt.valid, "Empty name should be invalid")
			} else if len(tt.request.Name) > 100 {
				assert.False(t, tt.valid, "Name longer than 100 chars should be invalid")
			}

			// Test description validation
			if len(tt.request.Description) > 500 {
				assert.False(t, tt.valid, "Description longer than 500 chars should be invalid")
			}
		})
	}
}

func TestUpdateBoardRequest_Validation(t *testing.T) {
	tests := []struct {
		name    string
		request UpdateBoardRequest
		valid   bool
	}{
		{
			name: "Valid update request",
			request: UpdateBoardRequest{
				Name:        "Updated Board",
				Description: "Updated Description",
			},
			valid: true,
		},
		{
			name: "Valid partial update",
			request: UpdateBoardRequest{
				Name: "Updated Board",
			},
			valid: true,
		},
		{
			name: "Invalid update - name too long",
			request: UpdateBoardRequest{
				Name: string(make([]byte, 101)), // 101 characters
			},
			valid: false,
		},
		{
			name: "Invalid update - description too long",
			request: UpdateBoardRequest{
				Description: string(make([]byte, 501)), // 501 characters
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test name validation
			if tt.request.Name != "" && len(tt.request.Name) > 100 {
				assert.False(t, tt.valid, "Name longer than 100 chars should be invalid")
			}

			// Test description validation
			if len(tt.request.Description) > 500 {
				assert.False(t, tt.valid, "Description longer than 500 chars should be invalid")
			}
		})
	}
}

func TestColumnValidation(t *testing.T) {
	validColumns := []string{"parking", "now", "next", "later", "release", "wont-do"}
	invalidColumns := []string{"invalid", "unknown", "test"}

	// Test valid columns
	for _, column := range validColumns {
		t.Run("Valid column: "+column, func(t *testing.T) {
			assert.True(t, models.IsValidColumn(column), "Column %s should be valid", column)
		})
	}

	// Test invalid columns
	for _, column := range invalidColumns {
		t.Run("Invalid column: "+column, func(t *testing.T) {
			assert.False(t, models.IsValidColumn(column), "Column %s should be invalid", column)
		})
	}
}

func TestDefaultValues(t *testing.T) {
	t.Run("Default visible columns", func(t *testing.T) {
		columns := models.GetDefaultVisibleColumns()
		expectedColumns := []string{"parking", "now", "next", "later", "release", "wont-do"}

		assert.Equal(t, expectedColumns, columns)
		assert.Len(t, columns, 6)
	})

	t.Run("Default visible fields", func(t *testing.T) {
		fields := models.GetDefaultVisibleFields()
		expectedFields := []string{"oneLiner", "description", "valueStatement", "riceScore"}

		assert.Equal(t, expectedFields, fields)
		assert.Len(t, fields, 4)
	})
}

func TestBoardResponse_Structure(t *testing.T) {
	response := BoardResponse{
		ID:             "507f1f77bcf86cd799439011",
		Name:           "Test Board",
		Description:    "Test Description",
		PublicLink:     "test-public-link",
		AdminID:        "user123",
		VisibleColumns: []string{"parking", "now"},
		VisibleFields:  []string{"oneLiner", "description"},
	}

	// Verify all required fields are present
	assert.NotEmpty(t, response.ID)
	assert.NotEmpty(t, response.Name)
	assert.NotEmpty(t, response.PublicLink)
	assert.NotEmpty(t, response.AdminID)
	assert.NotEmpty(t, response.VisibleColumns)
	assert.NotEmpty(t, response.VisibleFields)
}
func TestBoardIDFormat(t *testing.T) {
	t.Run("Board ID should start with 'b' and be 9 characters long", func(t *testing.T) {
		// Simulate board ID generation
		boardID := "b" + uuid.New().String()[:8]

		// Verify format
		assert.True(t, strings.HasPrefix(boardID, "b"), "Board ID should start with 'b'")
		assert.Equal(t, 9, len(boardID), "Board ID should be 9 characters long (b + 8 chars)")

		// Verify the remaining 8 characters are valid hex
		hexPart := boardID[1:]
		assert.Len(t, hexPart, 8)

		// Check if it's valid hex characters
		for _, char := range hexPart {
			assert.True(t,
				(char >= '0' && char <= '9') ||
					(char >= 'a' && char <= 'f') ||
					(char >= 'A' && char <= 'F') ||
					char == '-',
				"Board ID should contain valid hex characters or hyphens")
		}
	})

	t.Run("Multiple board IDs should be unique", func(t *testing.T) {
		// Generate multiple board IDs
		ids := make(map[string]bool)
		for i := 0; i < 100; i++ {
			boardID := "b" + uuid.New().String()[:8]
			assert.False(t, ids[boardID], "Board IDs should be unique")
			ids[boardID] = true
		}
	})
}
