package domain

// Importer defines the interface for reading snippets from an input file.
type Importer interface {
	Read(string) ([]*Snippet, error)
}
