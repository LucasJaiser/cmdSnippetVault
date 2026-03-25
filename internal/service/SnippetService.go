package service

import (
	"context"
	"lucasjaiser/goSnipperVault/internal/domain"
)

type SnippetService struct {
	repo      domain.SnippetRepository
	clipboard domain.Clipboard
}

func NewSnippetService(repo domain.SnippetRepository, clipboard domain.Clipboard) *SnippetService {

	return &SnippetService{
		repo:      repo,
		clipboard: clipboard,
	}
}

func (s *SnippetService) Create(ctx context.Context, snippet domain.Snippet) error {
	err := snippet.Validate()

	if err != nil {
		return err
	}

	err = s.repo.Create(ctx, &snippet)

	if err != nil {
		return err
	}

	return nil
}

func (s *SnippetService) Get(ctx context.Context, id int64) (*domain.Snippet, error) {

	snippet, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	snippet.UseCount += 1

	err = s.repo.Update(ctx, snippet)

	if err != nil {
		return nil, err
	}

	return snippet, err
}

func (s *SnippetService) List(ctx context.Context, filter domain.ListFilter) ([]*domain.Snippet, error) {

	snippetList, err := s.repo.List(ctx, filter)

	return snippetList, err

}

func (s *SnippetService) Update(ctx context.Context, snippet domain.Snippet) error {

	err := s.repo.Update(ctx, &snippet)

	return err
}

func (s *SnippetService) Delete(ctx context.Context, id int64) error {
	err := s.repo.Delete(ctx, id)
	return err
}

func (s *SnippetService) Search(ctx context.Context, query string) ([]*domain.Snippet, error) {

	snippetList, err := s.repo.Search(ctx, query)

	return snippetList, err
}

func (s *SnippetService) GetAndCopy(ctx context.Context, id int64) error {
	snippet, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	snippet.UseCount += 1

	err = s.repo.Update(ctx, snippet)

	if err != nil {
		return err
	}

	//copy To clipboard

	return err
}
func (s *SnippetService) ListTags(ctx context.Context) ([]*domain.TagWithCount, error) {
	tags, err := s.repo.ListTags(ctx)

	return tags, err
}
