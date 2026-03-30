package service

import (
	"context"
	"fmt"
	"lucasjaiser/goSnippetVault/internal/domain"
)

// SnippetService provides business logic for managing snippets.
type SnippetService struct {
	repo      domain.SnippetRepository
	clipboard domain.Clipboard
}

// NewSnippetService creates a new SnippetService with the given repository and clipboard.
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

	fmt.Printf("Snippet created with ID: %d\n", snippet.ID)

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

	if s.clipboard != nil && s.clipboard.IsAvailable() {

		if err := s.clipboard.Copy(snippet.Command); err != nil {
			return fmt.Errorf("could not copy to clipboard: %w", err)
		}
		fmt.Println("Copied to Clipboard")
	} else {
		fmt.Println("Clippboard unavailable")
	}

	return nil
}

func (s *SnippetService) ListTags(ctx context.Context) ([]*domain.TagWithCount, error) {
	tags, err := s.repo.ListTags(ctx)

	return tags, err
}

func (s *SnippetService) CreateBatch(ctx context.Context, snippets []*domain.Snippet, dryrun bool) (*domain.ImportStatistics, error) {
	var rejected int
	var clearToImport []*domain.Snippet

	//Validate
	for _, snippet := range snippets {
		err := snippet.Validate()

		if err != nil {
			rejected += 1
			continue
		}

		clearToImport = append(clearToImport, snippet)
	}

	stats := &domain.ImportStatistics{
		Created:    len(clearToImport),
		Duplicates: 0,
	}

	// Execute Batch on validated snippets
	if !dryrun {

		returnedStats, err := s.repo.CreateBatch(ctx, clearToImport)
		if err != nil {
			return nil, err
		}

		stats = returnedStats
	}

	stats.Rejected = rejected

	return stats, nil

}
