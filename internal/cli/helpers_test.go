package cli

import (
	"fmt"
	"testing"

	"lucasjaiser/goSnipperVault/internal/importer"

	"github.com/stretchr/testify/assert"
)

func TestGetImportForFileType(t *testing.T) {
	tests := []struct {
		name           string
		filename       string
		formatOverride string
		wantNil        bool
		wantType       any
	}{
		{
			name:           "override json ignores file extension",
			filename:       "snippets.yaml",
			formatOverride: "json",
			wantType:       &importer.JSONImporter{},
		},
		{
			name:           "override yaml ignores file extension",
			filename:       "snippets.json",
			formatOverride: "yaml",
			wantType:       &importer.YAMLImporter{},
		},
		{
			name:           "override yml ignores file extension",
			filename:       "snippets.json",
			formatOverride: "yml",
			wantType:       &importer.YAMLImporter{},
		},
		{
			name:           "override with unsupported format returns nil",
			filename:       "snippets.json",
			formatOverride: "xml",
			wantNil:        true,
		},
		{
			name:     "detects json from file extension",
			filename: "snippets.json",
			wantType: &importer.JSONImporter{},
		},
		{
			name:     "detects yaml from file extension",
			filename: "snippets.yaml",
			wantType: &importer.YAMLImporter{},
		},
		{
			name:     "detects yml from file extension",
			filename: "snippets.yml",
			wantType: &importer.YAMLImporter{},
		},
		{
			name:     "detects json from path with directories",
			filename: "/home/user/exports/snippets.json",
			wantType: &importer.JSONImporter{},
		},
		{
			name:    "unsupported extension returns nil",
			filename: "snippets.xml",
			wantNil: true,
		},
		{
			name:    "no extension returns nil",
			filename: "snippets",
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getImportForFileType(tt.filename, tt.formatOverride)

			if tt.wantNil {
				assert.Nil(t, result)
				return
			}

			assert.NotNil(t, result)
			assert.IsType(t, tt.wantType, result, fmt.Sprintf("expected %T, got %T", tt.wantType, result))
		})
	}
}
