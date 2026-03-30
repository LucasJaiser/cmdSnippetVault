package exporter

import (
	"encoding/json"
	"fmt"
	"lucasjaiser/goSnippetVault/internal/domain"
	"os"
)

// JSONExporter implements domain.Exporter for JSON output.
type JSONExporter struct {
}

// NewJSONExporter creates a new JSONExporter.
func NewJSONExporter() *JSONExporter {
	return &JSONExporter{}
}

func (e *JSONExporter) Write(snippets []*domain.Snippet, output string) error {

	bytes, err := json.Marshal(snippets)

	if err != nil {
		return fmt.Errorf("could not convert snippets to JSON: %w", err)
	}

	err = os.WriteFile(output, bytes, 0644)

	if err != nil {
		return fmt.Errorf("could not write file: %w", err)
	}

	return nil
}
