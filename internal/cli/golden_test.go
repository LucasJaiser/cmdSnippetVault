package cli

import (
	"lucasjaiser/goSnipperVault/internal/domain"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testTime = time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

func TestGetCommand_PrintSnippet(t *testing.T) {
	tests := []struct {
		name    string
		snippet *domain.Snippet
	}{
		{
			name: "full_snippet",
			snippet: &domain.Snippet{
				ID:          1,
				Command:     "docker ps -a",
				Description: "Show all containers",
				Tags:        []string{"docker", "devops"},
				UseCount:    5,
				CreatedAt:   testTime,
				UpdatedAt:   testTime,
			},
		},
		{
			name: "no_description",
			snippet: &domain.Snippet{
				ID:        2,
				Command:   "ls -la",
				Tags:      []string{"shell"},
				UseCount:  0,
				CreatedAt: testTime,
				UpdatedAt: testTime,
			},
		},
		{
			name: "no_tags",
			snippet: &domain.Snippet{
				ID:          3,
				Command:     "echo hello",
				Description: "prints hello",
				UseCount:    10,
				CreatedAt:   testTime,
				UpdatedAt:   testTime,
			},
		},
		{
			name: "minimal",
			snippet: &domain.Snippet{
				ID:        4,
				Command:   "pwd",
				CreatedAt: testTime,
				UpdatedAt: testTime,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := captureOutput(t, func() {
				GetCommand_PrintSnippet(tt.snippet)
			})

			expected := goldenFile(t, "get_"+tt.name, actual)
			assert.Equal(t, expected, actual)
		})
	}
}

func TestSearchCommand_PrintSnippets(t *testing.T) {
	tests := []struct {
		name     string
		snippets []*domain.Snippet
		limit    int
		json     bool
		pretty   bool
	}{
		{
			name: "multiple_snippets",
			snippets: []*domain.Snippet{
				{
					ID:          1,
					Command:     "docker ps -a",
					Description: "Show all containers",
					Tags:        []string{"docker"},
					UseCount:    5,
				},
				{
					ID:          2,
					Command:     "docker logs -f",
					Description: "Follow container logs",
					Tags:        []string{"docker", "logs"},
					UseCount:    3,
				},
			},
			limit: 20,
		},
		{
			name:     "empty_results",
			snippets: []*domain.Snippet{},
			limit:    20,
		},
		{
			name: "single_snippet",
			snippets: []*domain.Snippet{
				{
					ID:       1,
					Command:  "git status",
					Tags:     []string{"git"},
					UseCount: 10,
				},
			},
			limit: 20,
		},
		{
			name: "with_limit",
			snippets: []*domain.Snippet{
				{ID: 1, Command: "echo one", UseCount: 1},
				{ID: 2, Command: "echo two", UseCount: 2},
				{ID: 3, Command: "echo three", UseCount: 3},
			},
			limit: 2,
		},
		{
			name: "json_output",
			snippets: []*domain.Snippet{
				{
					ID:          1,
					Command:     "ls -la",
					Description: "list files",
					Tags:        []string{"shell"},
					UseCount:    2,
					CreatedAt:   testTime,
					UpdatedAt:   testTime,
				},
			},
			limit: 20,
			json:  true,
		},
		{
			name: "json_pretty_output",
			snippets: []*domain.Snippet{
				{
					ID:          1,
					Command:     "ls -la",
					Description: "list files",
					Tags:        []string{"shell"},
					UseCount:    2,
					CreatedAt:   testTime,
					UpdatedAt:   testTime,
				},
			},
			limit:  20,
			json:   true,
			pretty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := captureOutput(t, func() {
				err := SearchCommand_PrintSnippets(tt.snippets, tt.limit, tt.json, tt.pretty)
				require.NoError(t, err)
			})

			expected := goldenFile(t, "search_"+tt.name, actual)
			assert.Equal(t, expected, actual)
		})
	}
}

func TestListTagsCommand_PrintTags(t *testing.T) {
	tests := []struct {
		name string
		tags []*domain.TagWithCount
	}{
		{
			name: "multiple_tags",
			tags: []*domain.TagWithCount{
				{Name: "docker", Count: 5},
				{Name: "git", Count: 3},
				{Name: "shell", Count: 1},
			},
		},
		{
			name: "empty",
			tags: []*domain.TagWithCount{},
		},
		{
			name: "single_tag",
			tags: []*domain.TagWithCount{
				{Name: "linux", Count: 10},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := captureOutput(t, func() {
				ListTagsCommand_PrintTags(tt.tags)
			})

			expected := goldenFile(t, "list_tags_"+tt.name, actual)
			assert.Equal(t, expected, actual)
		})
	}
}
