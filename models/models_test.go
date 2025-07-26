package models

import (
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func TestBoardValidation(t *testing.T) {
	// Test valid board
	board := &Board{
		Name:           "Test Board",
		Description:    "A test board",
		PublicLink:     "test-public-link-123",
		AdminID:        "admin-123",
		VisibleColumns: GetDefaultVisibleColumns(),
		VisibleFields:  GetDefaultVisibleFields(),
	}

	errors := ValidateBoard(board)
	if len(errors) != 0 {
		t.Errorf("Expected no validation errors for valid board, got: %v", errors)
	}

	// Test invalid board - missing name
	invalidBoard := &Board{
		PublicLink: "test-link",
		AdminID:    "admin-123",
	}

	errors = ValidateBoard(invalidBoard)
	if len(errors) == 0 {
		t.Error("Expected validation errors for board with missing name")
	}

	// Check if name error is present
	found := false
	for _, err := range errors {
		if err.Field == "name" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected name validation error")
	}
}

func TestIdeaValidation(t *testing.T) {
	// Test valid idea
	idea := &Idea{
		BoardID:        bson.NewObjectID(),
		OneLiner:       "Test idea",
		Description:    "A test idea description",
		ValueStatement: "This provides value",
		RiceScore: RICEScore{
			Reach:      80,
			Impact:     70,
			Confidence: 4,
			Effort:     60,
		},
		Column:   string(ColumnParking),
		Position: 0,
		Status:   string(StatusDraft),
	}

	errors := ValidateIdea(idea)
	if len(errors) != 0 {
		t.Errorf("Expected no validation errors for valid idea, got: %v", errors)
	}

	// Test invalid idea - missing required fields
	invalidIdea := &Idea{
		BoardID: bson.NilObjectID,
	}

	errors = ValidateIdea(invalidIdea)
	if len(errors) == 0 {
		t.Error("Expected validation errors for idea with missing required fields")
	}
}

func TestRICEScoreValidation(t *testing.T) {
	// Test valid RICE score
	validScore := &RICEScore{
		Reach:      50,
		Impact:     75,
		Confidence: 2,
		Effort:     40,
	}

	if !validScore.IsValidRICEScore() {
		t.Error("Expected valid RICE score to pass validation")
	}

	// Test invalid RICE score - invalid confidence
	invalidScore := &RICEScore{
		Reach:      50,
		Impact:     75,
		Confidence: 3, // Invalid confidence value
		Effort:     40,
	}

	if invalidScore.IsValidRICEScore() {
		t.Error("Expected invalid RICE score to fail validation")
	}

	// Test RICE score calculation
	score := &RICEScore{
		Reach:      100,
		Impact:     80,
		Confidence: 4,
		Effort:     50,
	}

	expected := float64(100*80*4) / float64(50)
	actual := score.CalculateRICEScore()

	if actual != expected {
		t.Errorf("Expected RICE score %f, got %f", expected, actual)
	}
}

func TestColumnValidation(t *testing.T) {
	validColumns := []string{
		string(ColumnParking),
		string(ColumnNow),
		string(ColumnNext),
		string(ColumnLater),
		string(ColumnRelease),
		string(ColumnWontDo),
	}

	for _, column := range validColumns {
		if !IsValidColumn(column) {
			t.Errorf("Expected column %s to be valid", column)
		}
	}

	// Test invalid column
	if IsValidColumn("invalid-column") {
		t.Error("Expected invalid column to fail validation")
	}
}

func TestStatusValidation(t *testing.T) {
	validStatuses := []string{
		string(StatusDraft),
		string(StatusActive),
		string(StatusDone),
		string(StatusArchived),
	}

	for _, status := range validStatuses {
		if !IsValidStatus(status) {
			t.Errorf("Expected status %s to be valid", status)
		}
	}

	// Test invalid status
	if IsValidStatus("invalid-status") {
		t.Error("Expected invalid status to fail validation")
	}
}

func TestDefaultValues(t *testing.T) {
	// Test default visible columns
	defaultColumns := GetDefaultVisibleColumns()
	expectedColumns := 6 // parking, now, next, later, release, wont-do

	if len(defaultColumns) != expectedColumns {
		t.Errorf("Expected %d default columns, got %d", expectedColumns, len(defaultColumns))
	}

	// Test default visible fields
	defaultFields := GetDefaultVisibleFields()
	expectedFields := 4 // oneLiner, description, valueStatement, riceScore

	if len(defaultFields) != expectedFields {
		t.Errorf("Expected %d default fields, got %d", expectedFields, len(defaultFields))
	}
}

func TestTimestampHandling(t *testing.T) {
	// Test board timestamp handling
	board := &Board{
		Name:       "Test Board",
		PublicLink: "test-link",
		AdminID:    "admin-123",
	}

	// Validate should set timestamps
	ValidateBoard(board)

	if board.CreatedAt.IsZero() {
		t.Error("Expected CreatedAt to be set during validation")
	}

	if board.UpdatedAt.IsZero() {
		t.Error("Expected UpdatedAt to be set during validation")
	}

	// Test idea timestamp handling
	idea := &Idea{
		BoardID:        bson.NewObjectID(),
		OneLiner:       "Test",
		Description:    "Test desc",
		ValueStatement: "Test value",
		RiceScore: RICEScore{
			Reach:      50,
			Impact:     50,
			Confidence: 2,
			Effort:     50,
		},
		Column: string(ColumnParking),
		Status: string(StatusDraft),
	}

	ValidateIdea(idea)

	if idea.CreatedAt.IsZero() {
		t.Error("Expected CreatedAt to be set during validation")
	}

	if idea.UpdatedAt.IsZero() {
		t.Error("Expected UpdatedAt to be set during validation")
	}
}
