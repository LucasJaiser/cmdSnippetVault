package exporter

import (
	"fmt"
	"lucasjaiser/goSnipperVault/internal/domain"
	"os"

	"gopkg.in/yaml.v3"
)

// YAMLExporter implements domain.Exporter for YAML output.
type YAMLExporter struct {
}

// NewYAMLExporter creates a new YAMLExporter.
func NewYAMLExporter() *YAMLExporter {
	return &YAMLExporter{}
}

func (e *YAMLExporter) Write(snippets []*domain.Snippet, output string) error {

	bytes, err := yaml.Marshal(snippets)

	if err != nil {
		return fmt.Errorf("could not marshal snippets to yaml: %w", err)
	}

	err = os.WriteFile(output, bytes, 0644)
	if err != nil {
		return fmt.Errorf("could not write file: %w", err)
	}

	return nil
}
