package cli

import (
	"fmt"
	"testing"

	"lucasjaiser/goSnippetVault/internal/exporter"
	"lucasjaiser/goSnippetVault/internal/importer"

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
			name:           "override with unsupported format defaults to json",
			filename:       "snippets.yaml",
			formatOverride: "xml",
			wantType:       &importer.JSONImporter{},
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
			name:     "unsupported extension defaults to json",
			filename: "snippets.xml",
			wantType: &importer.JSONImporter{},
		},
		{
			name:     "no extension defaults to json",
			filename: "snippets",
			wantType: &importer.JSONImporter{},
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

func TestGetExporterForType(t *testing.T) {
	tests := []struct {
		name     string
		format   string
		wantNil  bool
		wantType any
	}{
		{
			name:     "json format",
			format:   "json",
			wantType: &exporter.JSONExporter{},
		},
		{
			name:     "json with dot prefix",
			format:   ".json",
			wantType: &exporter.JSONExporter{},
		},
		{
			name:     "yaml format",
			format:   "yaml",
			wantType: &exporter.YAMLExporter{},
		},
		{
			name:     "yaml with dot prefix",
			format:   ".yaml",
			wantType: &exporter.YAMLExporter{},
		},
		{
			name:     "yml format",
			format:   "yml",
			wantType: &exporter.YAMLExporter{},
		},
		{
			name:     "yml with dot prefix",
			format:   ".yml",
			wantType: &exporter.YAMLExporter{},
		},
		{
			name:     "unsupported format defaults to json",
			format:   "xml",
			wantType: &exporter.JSONExporter{},
		},
		{
			name:     "empty format defaults to json",
			format:   "",
			wantType: &exporter.JSONExporter{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getExporterForType(tt.format)

			if tt.wantNil {
				assert.Nil(t, result)
				return
			}

			assert.NotNil(t, result)
			assert.IsType(t, tt.wantType, result, fmt.Sprintf("expected %T, got %T", tt.wantType, result))
		})
	}
}
