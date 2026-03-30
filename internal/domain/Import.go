package domain

type Importer interface {
	Read(string) ([]*Snippet, error)
}
