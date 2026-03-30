package importer

import (
	"lucasjaiser/goSnippetVault/internal/domain"
	"os"

	"gopkg.in/yaml.v3"
)

// YAMLImporter implements domain.Importer for YAML files.
type YAMLImporter struct {
}

// NewYAMLImporter creates a new YAMLImporter.
func NewYAMLImporter() *YAMLImporter {
	return &YAMLImporter{}
}

func (y *YAMLImporter) Read(filename string) ([]*domain.Snippet, error) {

	fileBytes, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var importedSnippets []*domain.Snippet
	err = yaml.Unmarshal(fileBytes, &importedSnippets)
	if err != nil {
		return nil, err
	}

	return importedSnippets, nil
}
