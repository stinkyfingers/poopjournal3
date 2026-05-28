package models

import (
	"time"

	"github.com/google/uuid"
)

type Food struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	Tags        []string  `json:"tags,omitempty"`
	Timestamp   time.Time `json:"timestamp"`
}

func NewFoodWithDateTime(userID, name, description, datetimeStr string, tags []string) *Food {
	now := time.Now()

	// Parse the datetime string (format: "2006-01-02T15:04")
	timestamp, err := time.Parse("2006-01-02T15:04", datetimeStr)
	if err != nil {
		// If parsing fails, use current time
		timestamp = now
	}

	// Validate and clean tags using the same validation as poop
	validatedTags := validateTags(tags)

	return &Food{
		ID:          uuid.New().String(),
		UserID:      userID,
		Name:        name,
		Description: description,
		Tags:        validatedTags,
		Timestamp:   timestamp,
	}
}
