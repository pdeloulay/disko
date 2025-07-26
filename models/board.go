package models

import (
	"time"
)

// Board represents a board document in MongoDB
type Board struct {
	ID             string    `bson:"_id,omitempty" json:"id"`
	Name           string    `bson:"name" json:"name" validate:"required,min=1,max=100"`
	Description    string    `bson:"description,omitempty" json:"description,omitempty" validate:"max=500"`
	PublicLink     string    `bson:"public_link" json:"publicLink" validate:"required"`
	AdminID        string    `bson:"admin_id" json:"adminId" validate:"required"`
	VisibleColumns []string  `bson:"visible_columns" json:"visibleColumns"`
	VisibleFields  []string  `bson:"visible_fields" json:"visibleFields"`
	CreatedAt      time.Time `bson:"created_at" json:"createdAt"`
	UpdatedAt      time.Time `bson:"updated_at" json:"updatedAt"`
}

// ColumnType represents the different columns available in a board
type ColumnType string

const (
	ColumnParking ColumnType = "parking"
	ColumnNow     ColumnType = "now"
	ColumnNext    ColumnType = "next"
	ColumnLater   ColumnType = "later"
	ColumnRelease ColumnType = "release"
	ColumnWontDo  ColumnType = "wont-do"
)

// IdeaField represents the different fields that can be visible for ideas
type IdeaField string

const (
	FieldOneLiner       IdeaField = "oneLiner"
	FieldDescription    IdeaField = "description"
	FieldValueStatement IdeaField = "valueStatement"
	FieldRiceScore      IdeaField = "riceScore"
)

// GetDefaultVisibleColumns returns the default visible columns for a new board
func GetDefaultVisibleColumns() []string {
	return []string{
		string(ColumnParking),
		string(ColumnNow),
		string(ColumnNext),
		string(ColumnLater),
		string(ColumnRelease),
		string(ColumnWontDo),
	}
}

// GetDefaultVisibleFields returns the default visible fields for ideas
func GetDefaultVisibleFields() []string {
	return []string{
		string(FieldOneLiner),
		string(FieldDescription),
		string(FieldValueStatement),
		string(FieldRiceScore),
	}
}

// IsValidColumn checks if a column type is valid
func IsValidColumn(column string) bool {
	validColumns := []string{
		string(ColumnParking),
		string(ColumnNow),
		string(ColumnNext),
		string(ColumnLater),
		string(ColumnRelease),
		string(ColumnWontDo),
	}

	for _, valid := range validColumns {
		if column == valid {
			return true
		}
	}
	return false
}
