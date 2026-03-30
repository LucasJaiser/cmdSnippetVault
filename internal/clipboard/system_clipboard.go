package clipboard

import "github.com/atotto/clipboard"

// SystemClipboard implements domain.Clipboard using the system clipboard.
type SystemClipboard struct {
}

// NewSystemClipboard creates a new SystemClipboard.
func NewSystemClipboard() *SystemClipboard {
	return &SystemClipboard{}
}

func (s *SystemClipboard) Copy(value string) error {
	return clipboard.WriteAll(value)
}

func (s *SystemClipboard) IsAvailable() bool {
	return !clipboard.Unsupported
}
