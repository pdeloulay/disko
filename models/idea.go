package models

import (
	"time"
)

// Idea represents an idea document in MongoDB
type Idea struct {
	ID             string          `bson:"_id,omitempty" json:"id"`
	BoardID        string          `bson:"board_id" json:"boardId" validate:"required"`
	OneLiner       string          `bson:"one_liner" json:"oneLiner" validate:"required,min=1,max=200"`
	Description    string          `bson:"description" json:"description" validate:"omitempty,max=1000"`
	ValueStatement string          `bson:"value_statement" json:"valueStatement" validate:"omitempty,max=500"`
	RiceScore      RICEScore       `bson:"rice_score" json:"riceScore" validate:"omitempty"`
	Column         string          `bson:"column" json:"column" validate:"required"`
	Position       int             `bson:"position" json:"position" validate:"min=0"`
	InProgress     bool            `bson:"in_progress" json:"inProgress"`
	Status         string          `bson:"status" json:"status" validate:"required"`
	ThumbsUp       int             `bson:"thumbs_up" json:"thumbsUp" validate:"min=0"`
	EmojiReactions []EmojiReaction `bson:"emoji_reactions" json:"emojiReactions"`
	CreatedAt      time.Time       `bson:"created_at" json:"createdAt"`
	UpdatedAt      time.Time       `bson:"updated_at" json:"updatedAt"`
}

// RICEScore represents the RICE scoring system for ideas
type RICEScore struct {
	Reach      int `bson:"reach" json:"reach" validate:"min=0,max=10"`           // 0-10 scale
	Impact     int `bson:"impact" json:"impact" validate:"min=0,max=10"`         // 0-10 scale
	Confidence int `bson:"confidence" json:"confidence" validate:"min=0,max=10"` // 0-10 scale
	Effort     int `bson:"effort" json:"effort" validate:"oneof=1 3 8 21"`       // 1, 3, 8, 21 (Low, Medium, High, Very High)
}

// EmojiReaction represents emoji feedback on ideas
type EmojiReaction struct {
	Emoji string `bson:"emoji" json:"emoji" validate:"required"`
	Count int    `bson:"count" json:"count" validate:"min=0"`
}

// IdeaStatus represents the different statuses an idea can have
type IdeaStatus string

const (
	StatusDraft    IdeaStatus = "draft"
	StatusActive   IdeaStatus = "active"
	StatusDone     IdeaStatus = "done"
	StatusArchived IdeaStatus = "archived"
)

// IsValidStatus checks if an idea status is valid
func IsValidStatus(status string) bool {
	validStatuses := []string{
		string(StatusDraft),
		string(StatusActive),
		string(StatusDone),
		string(StatusArchived),
	}

	for _, valid := range validStatuses {
		if status == valid {
			return true
		}
	}
	return false
}

// CalculateRICEScore calculates the total RICE score
func (r *RICEScore) CalculateRICEScore() float64 {
	if r.Effort == 0 {
		return 0
	}
	// Use 0-10 scale directly (no need to convert from percentages)
	reach := float64(r.Reach)
	impact := float64(r.Impact)
	confidence := float64(r.Confidence)
	return (reach * impact * confidence) / float64(r.Effort)
}

// IsValidRICEScore validates the RICE score values
func (r *RICEScore) IsValidRICEScore() bool {
	if r.Reach < 0 || r.Reach > 10 {
		return false
	}
	if r.Impact < 0 || r.Impact > 10 {
		return false
	}
	if r.Confidence < 0 || r.Confidence > 10 {
		return false
	}
	if r.Effort != 1 && r.Effort != 3 && r.Effort != 8 && r.Effort != 21 {
		return false
	}
	return true
}
