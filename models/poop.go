package models

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

type Poop struct {
	ID           string    `json:"id"`
	UserID       string    `json:"user_id"`
	BristolScale int       `json:"bristol_scale"` // 1-7
	Urgency      int       `json:"urgency"`       // 1-10 scale
	Tags         []string  `json:"tags,omitempty"`
	Notes        string    `json:"notes,omitempty"`
	Timestamp    time.Time `json:"timestamp"`
}

func NewPoopWithDateTime(userID string, bristolScale, urgency int, notes, datetimeStr string, tags []string) *Poop {
	now := time.Now()

	// Parse the datetime string (format: "2006-01-02T15:04")
	timestamp, err := time.Parse("2006-01-02T15:04", datetimeStr)
	if err != nil {
		// If parsing fails, use current time
		timestamp = now
	}

	// Validate urgency (1-10)
	if urgency < 1 || urgency > 10 {
		urgency = 5 // Default to middle value if invalid
	}

	// Validate and clean tags
	validatedTags := validateTags(tags)

	return &Poop{
		ID:           uuid.New().String(),
		UserID:       userID,
		BristolScale: bristolScale,
		Urgency:      urgency,
		Tags:         validatedTags,
		Notes:        notes,
		Timestamp:    timestamp,
	}
}

// validateTags validates and cleans tag input
// Max 5 tags, each max 20 characters, no duplicates
func validateTags(tags []string) []string {
	if len(tags) == 0 {
		return []string{}
	}

	seen := make(map[string]bool)
	validated := make([]string, 0)

	for _, tag := range tags {
		// Clean the tag: trim whitespace and convert to lowercase
		cleaned := strings.TrimSpace(strings.ToLower(tag))

		// Skip empty tags
		if cleaned == "" {
			continue
		}

		// Limit tag length to 20 characters
		if len(cleaned) > 20 {
			cleaned = cleaned[:20]
		}

		// Skip duplicates
		if seen[cleaned] {
			continue
		}

		seen[cleaned] = true
		validated = append(validated, cleaned)

		// Limit to 5 tags
		if len(validated) >= 5 {
			break
		}
	}

	return validated
}
