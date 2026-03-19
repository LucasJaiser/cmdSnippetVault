package service

import (
	"context"
	"lucasjaiser/goSnipperVault/internal/domain"
)

type SnippetRepository interface {
	Create(ctx context.Context, snippet *domain.Snippet) error
	GetByID(ctx context.Context, id int64) (*domain.Snippet, error)
	List(ctx context.Context, filter domain.ListFilter) ([]domain.Snippet, error)
	Update(ctx context.Context, snippet *domain.Snippet) error
	Delete(ctx context.Context, id int64) error
	Search(ctx context.Context, query string) ([]domain.Snippet, error)
}
