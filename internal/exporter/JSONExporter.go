package exporter

import (
	"encoding/json"
	"fmt"
	"lucasjaiser/goSnipperVault/internal/domain"
	"os"
)

type JSONExporter struct {
}

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
