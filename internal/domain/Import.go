package domain

type Import interface {
	Read(string) ([]*Snippet, error)
}
