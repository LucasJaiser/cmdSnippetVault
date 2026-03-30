package domain

import (
	"fmt"
	"slices"
	"strings"
	"time"
)

// Snippet represents a saved shell command with metadata.
type Snippet struct {
	ID          int64
	Command     string
	Description string
	Tags        []string
	UseCount    int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// ListFilter holds filtering and pagination options for listing snippets.
type ListFilter struct {
	Tag    string
	Limit  int
	Offset int
}

// Validate checks that the snippet has a non-empty command and valid tags.
func (s *Snippet) Validate() error {

	if s.Command == "" {
		return &ValidationError{
			Field:   "Command",
			Message: "Command is Empty",
		}
	}

	uniquieList := []string{}
	for _, ss := range s.Tags {
		if strings.ToLower(ss) != ss {
			return &ValidationError{
				Field:   "Tag",
				Message: fmt.Sprintf("Tags have to be lowercase: %s", ss),
			}
		}

		if !slices.Contains(uniquieList, ss) {
			uniquieList = append(uniquieList, strings.ToLower(ss))
		} else {
			return &ValidationError{
				Field:   "Tag",
				Message: fmt.Sprintf("Duplicate Tag: %s", ss),
			}
		}
	}

	return nil
}
