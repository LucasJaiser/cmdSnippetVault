package service

import (
	"context"
	"errors"
	"lucasjaiser/goSnipperVault/internal/config"
	"lucasjaiser/goSnipperVault/internal/domain"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockRepository is a hand-written mock implementing domain.SnippetRepository.
type mockRepository struct {
	createFn   func(ctx context.Context, snippet *domain.Snippet) error
	getByIDFn  func(ctx context.Context, id int64) (*domain.Snippet, error)
	listFn     func(ctx context.Context, filter domain.ListFilter) ([]*domain.Snippet, error)
	updateFn   func(ctx context.Context, snippet *domain.Snippet) error
	deleteFn   func(ctx context.Context, id int64) error
	searchFn   func(ctx context.Context, query string) ([]*domain.Snippet, error)
	listTagsFn func(ctx context.Context) ([]*domain.TagWithCount, error)
}

func (m *mockRepository) Create(ctx context.Context, snippet *domain.Snippet) error {
	return m.createFn(ctx, snippet)
}

func (m *mockRepository) GetByID(ctx context.Context, id int64) (*domain.Snippet, error) {
	return m.getByIDFn(ctx, id)
}

func (m *mockRepository) List(ctx context.Context, filter domain.ListFilter) ([]*domain.Snippet, error) {
	return m.listFn(ctx, filter)
}

func (m *mockRepository) Update(ctx context.Context, snippet *domain.Snippet) error {
	return m.updateFn(ctx, snippet)
}

func (m *mockRepository) Delete(ctx context.Context, id int64) error {
	return m.deleteFn(ctx, id)
}

func (m *mockRepository) Search(ctx context.Context, query string) ([]*domain.Snippet, error) {
	return m.searchFn(ctx, query)
}

func (m *mockRepository) ListTags(ctx context.Context) ([]*domain.TagWithCount, error) {
	return m.listTagsFn(ctx)
}

// mockClipboard is a hand-written mock implementing domain.Clipboard.
type mockClipboard struct {
	copyFn        func(text string) error
	isAvailableFn func() bool
}

func (m *mockClipboard) Copy(text string) error {
	return m.copyFn(text)
}

func (m *mockClipboard) IsAvailable() bool {
	return m.isAvailableFn()
}

func newTestService(repo domain.SnippetRepository) *SnippetService {
	return NewSnippetService(repo, &mockClipboard{
		copyFn:        func(_ string) error { return nil },
		isAvailableFn: func() bool { return true },
	}, &config.Config{})
}

func TestSnippetService_Create(t *testing.T) {
	tests := []struct {
		name      string
		snippet   domain.Snippet
		createFn  func(ctx context.Context, snippet *domain.Snippet) error
		wantErr   bool
		errTarget error
	}{
		{
			name: "success",
			snippet: domain.Snippet{
				Command:     "echo hello",
				Description: "prints hello",
				Tags:        []string{"shell"},
			},
			createFn: func(_ context.Context, _ *domain.Snippet) error {
				return nil
			},
			wantErr: false,
		},
		{
			name: "validation error empty command",
			snippet: domain.Snippet{
				Command: "",
			},
			createFn: func(_ context.Context, _ *domain.Snippet) error {
				return nil
			},
			wantErr:   true,
			errTarget: domain.ErrValidation,
		},
		{
			name: "validation error uppercase tag",
			snippet: domain.Snippet{
				Command: "ls -la",
				Tags:    []string{"Shell"},
			},
			createFn: func(_ context.Context, _ *domain.Snippet) error {
				return nil
			},
			wantErr:   true,
			errTarget: domain.ErrValidation,
		},
		{
			name: "validation error duplicate tag",
			snippet: domain.Snippet{
				Command: "ls -la",
				Tags:    []string{"shell", "shell"},
			},
			createFn: func(_ context.Context, _ *domain.Snippet) error {
				return nil
			},
			wantErr:   true,
			errTarget: domain.ErrValidation,
		},
		{
			name: "repository error",
			snippet: domain.Snippet{
				Command:     "echo hello",
				Description: "prints hello",
			},
			createFn: func(_ context.Context, _ *domain.Snippet) error {
				return errors.New("db error")
			},
			wantErr: true,
		},
		{
			name: "duplicate entry error",
			snippet: domain.Snippet{
				Command: "echo hello",
			},
			createFn: func(_ context.Context, _ *domain.Snippet) error {
				return domain.ErrDuplicate
			},
			wantErr:   true,
			errTarget: domain.ErrDuplicate,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newTestService(&mockRepository{
				createFn: tt.createFn,
			})

			err := svc.Create(context.Background(), tt.snippet)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errTarget != nil {
					assert.True(t, errors.Is(err, tt.errTarget), "expected error %v, got %v", tt.errTarget, err)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestSnippetService_Get(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		id       int64
		getByID  func(ctx context.Context, id int64) (*domain.Snippet, error)
		updateFn func(ctx context.Context, snippet *domain.Snippet) error
		wantErr  bool
		wantUse  int
	}{
		{
			name: "success increments use count",
			id:   1,
			getByID: func(_ context.Context, _ int64) (*domain.Snippet, error) {
				return &domain.Snippet{
					ID:        1,
					Command:   "echo hello",
					UseCount:  5,
					CreatedAt: now,
				}, nil
			},
			updateFn: func(_ context.Context, _ *domain.Snippet) error {
				return nil
			},
			wantErr: false,
			wantUse: 6,
		},
		{
			name: "not found",
			id:   999,
			getByID: func(_ context.Context, _ int64) (*domain.Snippet, error) {
				return nil, domain.ErrNotFound
			},
			updateFn: func(_ context.Context, _ *domain.Snippet) error {
				return nil
			},
			wantErr: true,
		},
		{
			name: "update fails after get",
			id:   1,
			getByID: func(_ context.Context, _ int64) (*domain.Snippet, error) {
				return &domain.Snippet{
					ID:       1,
					Command:  "echo hello",
					UseCount: 0,
				}, nil
			},
			updateFn: func(_ context.Context, _ *domain.Snippet) error {
				return errors.New("update failed")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newTestService(&mockRepository{
				getByIDFn: tt.getByID,
				updateFn:  tt.updateFn,
			})

			snippet, err := svc.Get(context.Background(), tt.id)

			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, snippet)
			} else {
				require.NoError(t, err)
				require.NotNil(t, snippet)
				assert.Equal(t, tt.wantUse, snippet.UseCount)
			}
		})
	}
}

func TestSnippetService_List(t *testing.T) {
	tests := []struct {
		name    string
		filter  domain.ListFilter
		listFn  func(ctx context.Context, filter domain.ListFilter) ([]*domain.Snippet, error)
		wantLen int
		wantErr bool
	}{
		{
			name:   "success returns snippets",
			filter: domain.ListFilter{Limit: 10},
			listFn: func(_ context.Context, _ domain.ListFilter) ([]*domain.Snippet, error) {
				return []*domain.Snippet{
					{ID: 1, Command: "echo one"},
					{ID: 2, Command: "echo two"},
				}, nil
			},
			wantLen: 2,
			wantErr: false,
		},
		{
			name:   "empty list",
			filter: domain.ListFilter{Tag: "nonexistent"},
			listFn: func(_ context.Context, _ domain.ListFilter) ([]*domain.Snippet, error) {
				return []*domain.Snippet{}, nil
			},
			wantLen: 0,
			wantErr: false,
		},
		{
			name:   "repository error",
			filter: domain.ListFilter{},
			listFn: func(_ context.Context, _ domain.ListFilter) ([]*domain.Snippet, error) {
				return nil, errors.New("db error")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newTestService(&mockRepository{
				listFn: tt.listFn,
			})

			snippets, err := svc.List(context.Background(), tt.filter)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Len(t, snippets, tt.wantLen)
			}
		})
	}
}

func TestSnippetService_Update(t *testing.T) {
	tests := []struct {
		name     string
		snippet  domain.Snippet
		updateFn func(ctx context.Context, snippet *domain.Snippet) error
		wantErr  bool
	}{
		{
			name: "success",
			snippet: domain.Snippet{
				ID:      1,
				Command: "echo updated",
			},
			updateFn: func(_ context.Context, _ *domain.Snippet) error {
				return nil
			},
			wantErr: false,
		},
		{
			name: "not found",
			snippet: domain.Snippet{
				ID:      999,
				Command: "echo missing",
			},
			updateFn: func(_ context.Context, _ *domain.Snippet) error {
				return domain.ErrNotFound
			},
			wantErr: true,
		},
		{
			name: "repository error",
			snippet: domain.Snippet{
				ID:      1,
				Command: "echo fail",
			},
			updateFn: func(_ context.Context, _ *domain.Snippet) error {
				return errors.New("db error")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newTestService(&mockRepository{
				updateFn: tt.updateFn,
			})

			err := svc.Update(context.Background(), tt.snippet)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestSnippetService_Delete(t *testing.T) {
	tests := []struct {
		name     string
		id       int64
		deleteFn func(ctx context.Context, id int64) error
		wantErr  bool
	}{
		{
			name: "success",
			id:   1,
			deleteFn: func(_ context.Context, _ int64) error {
				return nil
			},
			wantErr: false,
		},
		{
			name: "not found",
			id:   999,
			deleteFn: func(_ context.Context, _ int64) error {
				return domain.ErrNotFound
			},
			wantErr: true,
		},
		{
			name: "repository error",
			id:   1,
			deleteFn: func(_ context.Context, _ int64) error {
				return errors.New("db error")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newTestService(&mockRepository{
				deleteFn: tt.deleteFn,
			})

			err := svc.Delete(context.Background(), tt.id)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestSnippetService_Search(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		searchFn func(ctx context.Context, query string) ([]*domain.Snippet, error)
		wantLen  int
		wantErr  bool
	}{
		{
			name:  "success with results",
			query: "echo",
			searchFn: func(_ context.Context, _ string) ([]*domain.Snippet, error) {
				return []*domain.Snippet{
					{ID: 1, Command: "echo hello"},
					{ID: 2, Command: "echo world"},
				}, nil
			},
			wantLen: 2,
			wantErr: false,
		},
		{
			name:  "no results",
			query: "nonexistent",
			searchFn: func(_ context.Context, _ string) ([]*domain.Snippet, error) {
				return []*domain.Snippet{}, nil
			},
			wantLen: 0,
			wantErr: false,
		},
		{
			name:  "repository error",
			query: "fail",
			searchFn: func(_ context.Context, _ string) ([]*domain.Snippet, error) {
				return nil, errors.New("db error")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newTestService(&mockRepository{
				searchFn: tt.searchFn,
			})

			snippets, err := svc.Search(context.Background(), tt.query)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Len(t, snippets, tt.wantLen)
			}
		})
	}
}

func TestSnippetService_GetAndCopy(t *testing.T) {
	tests := []struct {
		name     string
		id       int64
		getByID  func(ctx context.Context, id int64) (*domain.Snippet, error)
		updateFn func(ctx context.Context, snippet *domain.Snippet) error
		wantErr  bool
	}{
		{
			name: "success",
			id:   1,
			getByID: func(_ context.Context, _ int64) (*domain.Snippet, error) {
				return &domain.Snippet{
					ID:       1,
					Command:  "echo hello",
					UseCount: 3,
				}, nil
			},
			updateFn: func(_ context.Context, s *domain.Snippet) error {
				assert.Equal(t, 4, s.UseCount)
				return nil
			},
			wantErr: false,
		},
		{
			name: "get not found",
			id:   999,
			getByID: func(_ context.Context, _ int64) (*domain.Snippet, error) {
				return nil, domain.ErrNotFound
			},
			updateFn: func(_ context.Context, _ *domain.Snippet) error {
				return nil
			},
			wantErr: true,
		},
		{
			name: "update fails",
			id:   1,
			getByID: func(_ context.Context, _ int64) (*domain.Snippet, error) {
				return &domain.Snippet{
					ID:       1,
					Command:  "echo hello",
					UseCount: 0,
				}, nil
			},
			updateFn: func(_ context.Context, _ *domain.Snippet) error {
				return errors.New("update failed")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newTestService(&mockRepository{
				getByIDFn: tt.getByID,
				updateFn:  tt.updateFn,
			})

			err := svc.GetAndCopy(context.Background(), tt.id)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestSnippetService_ListTags(t *testing.T) {
	tests := []struct {
		name       string
		listTagsFn func(ctx context.Context) ([]*domain.TagWithCount, error)
		wantLen    int
		wantErr    bool
	}{
		{
			name: "success",
			listTagsFn: func(_ context.Context) ([]*domain.TagWithCount, error) {
				return []*domain.TagWithCount{
					{Name: "shell", Count: 5},
					{Name: "git", Count: 3},
				}, nil
			},
			wantLen: 2,
			wantErr: false,
		},
		{
			name: "empty",
			listTagsFn: func(_ context.Context) ([]*domain.TagWithCount, error) {
				return []*domain.TagWithCount{}, nil
			},
			wantLen: 0,
			wantErr: false,
		},
		{
			name: "repository error",
			listTagsFn: func(_ context.Context) ([]*domain.TagWithCount, error) {
				return nil, errors.New("db error")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newTestService(&mockRepository{
				listTagsFn: tt.listTagsFn,
			})

			tags, err := svc.ListTags(context.Background())

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Len(t, tags, tt.wantLen)
			}
		})
	}
}
