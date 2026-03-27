package clipboard

import (
	"fmt"
)

type NoopClipboard struct {
}

func NewNoopClipboard() *NoopClipboard {
	return &NoopClipboard{}
}

func (n *NoopClipboard) Copy(value string) error {
	return fmt.Errorf("clipboard unavailable")
}

func (n *NoopClipboard) IsAvailable() bool {
	return false
}
