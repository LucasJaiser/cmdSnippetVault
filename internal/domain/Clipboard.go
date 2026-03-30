package domain

// Clipboard provides an abstraction for copying text to the system clipboard.
type Clipboard interface {
	Copy(string) error
	IsAvailable() bool
}
