package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSnippetValidation(t *testing.T) {
	tests := []struct {
		name    string
		snippet Snippet
		wantErr error
	}{
		{
			name:    "valid Snippet",
			snippet: Snippet{Command: "ls -la", Tags: []string{"linux"}},
			wantErr: nil,
		},
		{
			name:    "invalid Snippet",
			snippet: Snippet{Command: ""},
			wantErr: ErrValidation,
		},
		{
			name:    "Non Lowercase tag",
			snippet: Snippet{Command: "ls -la", Tags: []string{"linux", "List"}},
			wantErr: ErrValidation,
		},
		{
			name:    "Duplicate Tag",
			snippet: Snippet{Command: "ls -la", Tags: []string{"linux", "linux"}},
			wantErr: ErrValidation,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.snippet.Validate()
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
