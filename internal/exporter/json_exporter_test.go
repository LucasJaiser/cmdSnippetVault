package exporter

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"lucasjaiser/goSnippetVault/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJSONExporter_Write(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		snippets []*domain.Snippet
		wantErr  bool
	}{
		{
			name: "writes multiple snippets",
			snippets: []*domain.Snippet{
				{ID: 1, Command: "ls -la", Description: "list files", Tags: []string{"filesystem"}, CreatedAt: now, UpdatedAt: now},
				{ID: 2, Command: "git status", Description: "check status", Tags: []string{"git"}, CreatedAt: now, UpdatedAt: now},
			},
		},
		{
			name:     "writes empty list",
			snippets: []*domain.Snippet{},
		},
		{
			name: "writes snippet without tags",
			snippets: []*domain.Snippet{
				{ID: 1, Command: "echo hello", Description: "greeting", CreatedAt: now, UpdatedAt: now},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			outPath := filepath.Join(t.TempDir(), "export.json")

			exp := NewJSONExporter()
			err := exp.Write(tt.snippets, outPath)

			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			data, err := os.ReadFile(outPath)
			require.NoError(t, err)

			var result []*domain.Snippet
			err = json.Unmarshal(data, &result)
			require.NoError(t, err)

			assert.Len(t, result, len(tt.snippets))
			for i, s := range result {
				assert.Equal(t, tt.snippets[i].Command, s.Command)
				assert.Equal(t, tt.snippets[i].Description, s.Description)
			}
		})
	}
}

func TestJSONExporter_Write_InvalidPath(t *testing.T) {
	exp := NewJSONExporter()
	err := exp.Write([]*domain.Snippet{{Command: "ls"}}, "/nonexistent/dir/file.json")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "could not write file")
}
