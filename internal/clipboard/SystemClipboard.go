package clipboard

import "github.com/atotto/clipboard"

type SystemClipboard struct {
}

func NewSystemClipboard() *SystemClipboard {
	return &SystemClipboard{}
}

func (s *SystemClipboard) Copy(value string) error {
	return clipboard.WriteAll(value)
}

func (s *SystemClipboard) IsAvailable() bool {
	return !clipboard.Unsupported
}
