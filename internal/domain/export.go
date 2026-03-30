package domain

// Exporter defines the interface for writing snippets to an output file.
type Exporter interface {
	Write([]*Snippet, string) error
}
