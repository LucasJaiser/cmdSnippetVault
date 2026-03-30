package importer

import (
	"encoding/json"
	"lucasjaiser/goSnipperVault/internal/domain"
	"os"
)

// JSONImporter implements domain.Importer for JSON files.
type JSONImporter struct {
}

// NewJSONImporter creates a new JSONImporter.
func NewJSONImporter() *JSONImporter {
	return &JSONImporter{}
}

func (j *JSONImporter) Read(filename string) ([]*domain.Snippet, error) {

	fileBytes, err := os.ReadFile(filename)

	if err != nil {
		return nil, err
	}

	var importedSnippets []*domain.Snippet
	err = json.Unmarshal(fileBytes, &importedSnippets)
	if err != nil {
		return nil, err
	}

	return importedSnippets, nil
}
