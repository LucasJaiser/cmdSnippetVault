package domain

import (
	"context"
)

type SnippetRepository interface {
	Create(ctx context.Context, snippet *Snippet) error
	GetByID(ctx context.Context, id int64) (*Snippet, error)
	List(ctx context.Context, filter ListFilter) ([]*Snippet, error)
	Update(ctx context.Context, snippet *Snippet) error
	Delete(ctx context.Context, id int64) error
	Search(ctx context.Context, query string) ([]*Snippet, error)
	ListTags(ctx context.Context) ([]*TagWithCount, error)
	CreateBatch(ctx context.Context, snippets []*Snippet) (*ImportStatistics, error)
}
