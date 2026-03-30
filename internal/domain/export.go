package domain

type Exporter interface {
	Write([]*Snippet, string) error
}
