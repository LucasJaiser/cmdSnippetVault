package domain

type Clipboard interface {
	Copy(string) error
	IsAvailable() bool
}
