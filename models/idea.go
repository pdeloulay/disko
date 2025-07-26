package models

import (
	"time"
)

// Idea represents an idea document in MongoDB
type Idea struct {
	ID             string          `bson:"_id,omitempty" json:"id"`
	BoardID        string          `bson:"board_id" json:"boardId" validate:"required"`
	OneLiner       string          `bson:"one_liner" json:"oneLiner" validate:"required,min=1,max=200"`
	Description    string          `bson:"description" json:"description" validate:"required,min=1,max=1000"`
	ValueStatement string          `bson:"value_statement" json:"valueStatement" validate:"required,min=1,max=500"`
	RiceScore      RICEScore       `bson:"rice_score" json:"riceScore" validate:"required"`
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
	Reach      int `bson:"reach" json:"reach" validate:"min=0,max=100"`           // 0-100%
	Impact     int `bson:"impact" json:"impact" validate:"min=0,max=100"`         // 0-100%
	Confidence int `bson:"confidence" json:"confidence" validate:"oneof=1 2 4 8"` // 1, 2, 4, 8 (hours, days, weeks, months)
	Effort     int `bson:"effort" json:"effort" validate:"min=0,max=100"`         // 0-100%
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
	return float64(r.Reach*r.Impact*r.Confidence) / float64(r.Effort)
}

// IsValidRICEScore validates the RICE score values
func (r *RICEScore) IsValidRICEScore() bool {
	if r.Reach < 0 || r.Reach > 100 {
		return false
	}
	if r.Impact < 0 || r.Impact > 100 {
		return false
	}
	if r.Confidence != 1 && r.Confidence != 2 && r.Confidence != 4 && r.Confidence != 8 {
		return false
	}
	if r.Effort < 0 || r.Effort > 100 {
		return false
	}
	return true
}
