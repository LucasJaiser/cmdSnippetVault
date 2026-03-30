package clipboard

import (
	"fmt"
)

// NoopClipboard implements domain.Clipboard as a no-op for environments without clipboard support.
type NoopClipboard struct {
}

// NewNoopClipboard creates a new NoopClipboard.
func NewNoopClipboard() *NoopClipboard {
	return &NoopClipboard{}
}

func (n *NoopClipboard) Copy(value string) error {
	return fmt.Errorf("clipboard unavailable")
}

func (n *NoopClipboard) IsAvailable() bool {
	return false
}
