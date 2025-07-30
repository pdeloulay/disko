package models

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

// ValidationError represents a validation error
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error on field '%s': %s", e.Field, e.Message)
}

// ValidationErrors represents multiple validation errors
type ValidationErrors []ValidationError

func (e ValidationErrors) Error() string {
	if len(e) == 0 {
		return ""
	}

	var messages []string
	for _, err := range e {
		messages = append(messages, err.Error())
	}
	return strings.Join(messages, "; ")
}

// ValidateBoard validates a Board struct
func ValidateBoard(board *Board) ValidationErrors {
	var errors ValidationErrors

	// Validate name
	if strings.TrimSpace(board.Name) == "" {
		errors = append(errors, ValidationError{
			Field:   "name",
			Message: "name is required",
		})
	} else if len(board.Name) > 100 {
		errors = append(errors, ValidationError{
			Field:   "name",
			Message: "name must be 100 characters or less",
		})
	}

	// Validate description length
	if len(board.Description) > 500 {
		errors = append(errors, ValidationError{
			Field:   "description",
			Message: "description must be 500 characters or less",
		})
	}

	// Validate public link
	if strings.TrimSpace(board.PublicLink) == "" {
		errors = append(errors, ValidationError{
			Field:   "publicLink",
			Message: "public link is required",
		})
	}

	// Validate user ID
	if strings.TrimSpace(board.UserID) == "" {
		errors = append(errors, ValidationError{
			Field:   "userId",
			Message: "user ID is required",
		})
	}

	// Validate visible columns
	for _, column := range board.VisibleColumns {
		if !IsValidColumn(column) {
			errors = append(errors, ValidationError{
				Field:   "visibleColumns",
				Message: fmt.Sprintf("invalid column type: %s", column),
			})
		}
	}

	// Set timestamps if not set
	if board.CreatedAt.IsZero() {
		board.CreatedAt = time.Now().UTC()
	}
	board.UpdatedAt = time.Now().UTC()

	return errors
}

// ValidateIdea validates an Idea struct
func ValidateIdea(idea *Idea) ValidationErrors {
	var errors ValidationErrors

	// Validate board ID
	if strings.TrimSpace(idea.BoardID) == "" {
		errors = append(errors, ValidationError{
			Field:   "boardId",
			Message: "board ID is required",
		})
	}

	// Validate one-liner
	if strings.TrimSpace(idea.OneLiner) == "" {
		errors = append(errors, ValidationError{
			Field:   "oneLiner",
			Message: "one-liner is required",
		})
	} else if len(idea.OneLiner) > 200 {
		errors = append(errors, ValidationError{
			Field:   "oneLiner",
			Message: "one-liner must be 200 characters or less",
		})
	}

	// Validate description
	if strings.TrimSpace(idea.Description) == "" {
		errors = append(errors, ValidationError{
			Field:   "description",
			Message: "description is required",
		})
	} else if len(idea.Description) > 1000 {
		errors = append(errors, ValidationError{
			Field:   "description",
			Message: "description must be 1000 characters or less",
		})
	}

	// Validate value statement
	if strings.TrimSpace(idea.ValueStatement) == "" {
		errors = append(errors, ValidationError{
			Field:   "valueStatement",
			Message: "value statement is required",
		})
	} else if len(idea.ValueStatement) > 500 {
		errors = append(errors, ValidationError{
			Field:   "valueStatement",
			Message: "value statement must be 500 characters or less",
		})
	}

	// Validate RICE score
	if !idea.RiceScore.IsValidRICEScore() {
		errors = append(errors, ValidationError{
			Field:   "riceScore",
			Message: "invalid RICE score values",
		})
	}

	// Validate column
	if !IsValidColumn(idea.Column) {
		errors = append(errors, ValidationError{
			Field:   "column",
			Message: fmt.Sprintf("invalid column type: %s", idea.Column),
		})
	}

	// Validate position
	if idea.Position < 0 {
		errors = append(errors, ValidationError{
			Field:   "position",
			Message: "position must be non-negative",
		})
	}

	// Validate status
	if !IsValidStatus(idea.Status) {
		errors = append(errors, ValidationError{
			Field:   "status",
			Message: fmt.Sprintf("invalid status: %s", idea.Status),
		})
	}

	// Validate thumbs up count
	if idea.ThumbsUp < 0 {
		errors = append(errors, ValidationError{
			Field:   "thumbsUp",
			Message: "thumbs up count must be non-negative",
		})
	}

	// Validate emoji reactions
	for i, reaction := range idea.EmojiReactions {
		if strings.TrimSpace(reaction.Emoji) == "" {
			errors = append(errors, ValidationError{
				Field:   fmt.Sprintf("emojiReactions[%d].emoji", i),
				Message: "emoji is required",
			})
		}
		if reaction.Count < 0 {
			errors = append(errors, ValidationError{
				Field:   fmt.Sprintf("emojiReactions[%d].count", i),
				Message: "emoji count must be non-negative",
			})
		}
	}

	// Set timestamps if not set
	if idea.CreatedAt.IsZero() {
		idea.CreatedAt = time.Now().UTC()
	}
	idea.UpdatedAt = time.Now().UTC()

	return errors
}

// IsValidUUID checks if a string is a valid UUID format
func IsValidUUID(uuid string) bool {
	uuidRegex := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
	return uuidRegex.MatchString(strings.ToLower(uuid))
}

// IsValidEmail checks if a string is a valid email format
func IsValidEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}
