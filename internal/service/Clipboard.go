package service

type Clipboard interface {
	Copy(test string) error
	IsAvailable() bool
}
