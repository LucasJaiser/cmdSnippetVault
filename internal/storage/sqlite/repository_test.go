package sqlite

import (
	"context"
	"errors"
	"lucasjaiser/goSnipperVault/internal/domain"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) *SQLiteRepository {
	t.Helper()
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	repo, err := New(dbPath)
	require.NoError(t, err, "failed to create test database")
	t.Cleanup(func() {
		repo.Close()
		os.Remove(dbPath)
	})
	return repo
}

func createTestSnippet(t *testing.T, repo *SQLiteRepository, command, description string, tags []string) *domain.Snippet {
	t.Helper()
	snippet := &domain.Snippet{
		Command:     command,
		Description: description,
		Tags:        tags,
	}
	err := repo.Create(context.Background(), snippet)
	require.NoError(t, err)
	require.NotZero(t, snippet.ID)
	return snippet
}

func TestCreate(t *testing.T) {
	tests := []struct {
		name        string
		snippet     *domain.Snippet
		wantErr     bool
		checkResult func(t *testing.T, repo *SQLiteRepository, snippet *domain.Snippet)
	}{
		{
			name: "creates snippet without tags",
			snippet: &domain.Snippet{
				Command:     "ls -la",
				Description: "list all files",
			},
			checkResult: func(t *testing.T, repo *SQLiteRepository, snippet *domain.Snippet) {
				assert.NotZero(t, snippet.ID)
				got, err := repo.GetByID(context.Background(), snippet.ID)
				require.NoError(t, err)
				assert.Equal(t, "ls -la", got.Command)
				assert.Equal(t, "list all files", got.Description)
				assert.Empty(t, got.Tags)
			},
		},
		{
			name: "creates snippet with tags",
			snippet: &domain.Snippet{
				Command:     "docker ps",
				Description: "list containers",
				Tags:        []string{"docker", "devops"},
			},
			checkResult: func(t *testing.T, repo *SQLiteRepository, snippet *domain.Snippet) {
				got, err := repo.GetByID(context.Background(), snippet.ID)
				require.NoError(t, err)
				assert.Equal(t, "docker ps", got.Command)
				assert.ElementsMatch(t, []string{"docker", "devops"}, got.Tags)
			},
		},
		{
			name: "reuses existing tags across snippets",
			snippet: &domain.Snippet{
				Command:     "docker compose up",
				Description: "start compose",
				Tags:        []string{"docker"},
			},
			checkResult: func(t *testing.T, repo *SQLiteRepository, snippet *domain.Snippet) {
				// Create another snippet with the same tag
				second := &domain.Snippet{
					Command:     "docker compose down",
					Description: "stop compose",
					Tags:        []string{"docker"},
				}
				err := repo.Create(context.Background(), second)
				require.NoError(t, err)

				got1, err := repo.GetByID(context.Background(), snippet.ID)
				require.NoError(t, err)
				got2, err := repo.GetByID(context.Background(), second.ID)
				require.NoError(t, err)
				assert.ElementsMatch(t, []string{"docker"}, got1.Tags)
				assert.ElementsMatch(t, []string{"docker"}, got2.Tags)
			},
		},
		{
			name: "sets ID on snippet after creation",
			snippet: &domain.Snippet{
				Command: "echo hello",
			},
			checkResult: func(t *testing.T, repo *SQLiteRepository, snippet *domain.Snippet) {
				assert.Greater(t, snippet.ID, int64(0))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := setupTestDB(t)
			err := repo.Create(context.Background(), tt.snippet)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			if tt.checkResult != nil {
				tt.checkResult(t, repo, tt.snippet)
			}
		})
	}
}

func TestGetByID(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(t *testing.T, repo *SQLiteRepository) int64
		wantErr error
	}{
		{
			name: "returns existing snippet",
			setup: func(t *testing.T, repo *SQLiteRepository) int64 {
				s := createTestSnippet(t, repo, "git status", "check status", []string{"git"})
				return s.ID
			},
		},
		{
			name: "returns ErrNotFound for nonexistent ID",
			setup: func(t *testing.T, repo *SQLiteRepository) int64 {
				return 9999
			},
			wantErr: domain.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := setupTestDB(t)
			id := tt.setup(t, repo)

			got, err := repo.GetByID(context.Background(), id)
			if tt.wantErr != nil {
				require.Error(t, err)
				assert.True(t, errors.Is(err, tt.wantErr))
				return
			}
			require.NoError(t, err)
			assert.Equal(t, id, got.ID)
			assert.Equal(t, "git status", got.Command)
			assert.Equal(t, "check status", got.Description)
			assert.ElementsMatch(t, []string{"git"}, got.Tags)
			assert.NotZero(t, got.CreatedAt)
			assert.NotZero(t, got.UpdatedAt)
			assert.Equal(t, 0, got.UseCount)
		})
	}
}

func TestDelete(t *testing.T) {
	tests := []struct {
		name  string
		setup func(t *testing.T, repo *SQLiteRepository) int64
	}{
		{
			name: "deletes existing snippet",
			setup: func(t *testing.T, repo *SQLiteRepository) int64 {
				s := createTestSnippet(t, repo, "rm -rf /tmp/test", "cleanup", []string{"cleanup"})
				return s.ID
			},
		},
		{
			name: "no error when deleting nonexistent ID",
			setup: func(t *testing.T, repo *SQLiteRepository) int64 {
				return 9999
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := setupTestDB(t)
			id := tt.setup(t, repo)

			err := repo.Delete(context.Background(), id)
			require.NoError(t, err)

			_, err = repo.GetByID(context.Background(), id)
			assert.True(t, errors.Is(err, domain.ErrNotFound))
		})
	}
}

func TestUpdate(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(t *testing.T, repo *SQLiteRepository) *domain.Snippet
		update  func(snippet *domain.Snippet)
		check   func(t *testing.T, repo *SQLiteRepository, snippet *domain.Snippet)
		wantErr bool
	}{
		{
			name: "updates command and description",
			setup: func(t *testing.T, repo *SQLiteRepository) *domain.Snippet {
				return createTestSnippet(t, repo, "old-cmd", "old desc", nil)
			},
			update: func(s *domain.Snippet) {
				s.Command = "new-cmd"
				s.Description = "new desc"
			},
			check: func(t *testing.T, repo *SQLiteRepository, s *domain.Snippet) {
				got, err := repo.GetByID(context.Background(), s.ID)
				require.NoError(t, err)
				assert.Equal(t, "new-cmd", got.Command)
				assert.Equal(t, "new desc", got.Description)
			},
		},
		{
			name: "updates tags replacing old ones",
			setup: func(t *testing.T, repo *SQLiteRepository) *domain.Snippet {
				return createTestSnippet(t, repo, "cmd", "desc", []string{"old-tag"})
			},
			update: func(s *domain.Snippet) {
				s.Tags = []string{"new-tag", "another-tag"}
			},
			check: func(t *testing.T, repo *SQLiteRepository, s *domain.Snippet) {
				got, err := repo.GetByID(context.Background(), s.ID)
				require.NoError(t, err)
				assert.ElementsMatch(t, []string{"new-tag", "another-tag"}, got.Tags)
			},
		},
		{
			name: "updates use count",
			setup: func(t *testing.T, repo *SQLiteRepository) *domain.Snippet {
				return createTestSnippet(t, repo, "cmd", "desc", nil)
			},
			update: func(s *domain.Snippet) {
				s.UseCount = 5
			},
			check: func(t *testing.T, repo *SQLiteRepository, s *domain.Snippet) {
				got, err := repo.GetByID(context.Background(), s.ID)
				require.NoError(t, err)
				assert.Equal(t, 5, got.UseCount)
			},
		},
		{
			name: "creates snippet when ID is zero",
			setup: func(t *testing.T, repo *SQLiteRepository) *domain.Snippet {
				return &domain.Snippet{
					Command:     "new-cmd",
					Description: "auto-created",
					Tags:        []string{"auto"},
				}
			},
			update: func(s *domain.Snippet) {},
			check: func(t *testing.T, repo *SQLiteRepository, s *domain.Snippet) {
				assert.NotZero(t, s.ID)
				got, err := repo.GetByID(context.Background(), s.ID)
				require.NoError(t, err)
				assert.Equal(t, "new-cmd", got.Command)
			},
		},
		{
			name: "creates snippet when ID does not exist",
			setup: func(t *testing.T, repo *SQLiteRepository) *domain.Snippet {
				return &domain.Snippet{
					ID:          9999,
					Command:     "ghost-cmd",
					Description: "nonexistent",
				}
			},
			update: func(s *domain.Snippet) {},
			check: func(t *testing.T, repo *SQLiteRepository, s *domain.Snippet) {
				assert.NotZero(t, s.ID)
				got, err := repo.GetByID(context.Background(), s.ID)
				require.NoError(t, err)
				assert.Equal(t, "ghost-cmd", got.Command)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := setupTestDB(t)
			snippet := tt.setup(t, repo)
			tt.update(snippet)

			err := repo.Update(context.Background(), snippet)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			if tt.check != nil {
				tt.check(t, repo, snippet)
			}
		})
	}
}

func TestList(t *testing.T) {
	tests := []struct {
		name   string
		setup  func(t *testing.T, repo *SQLiteRepository)
		filter domain.ListFilter
		check  func(t *testing.T, snippets []domain.Snippet)
	}{
		{
			name:   "returns empty list when no snippets exist",
			setup:  func(t *testing.T, repo *SQLiteRepository) {},
			filter: domain.ListFilter{},
			check: func(t *testing.T, snippets []domain.Snippet) {
				assert.Empty(t, snippets)
			},
		},
		{
			name: "returns all snippets",
			setup: func(t *testing.T, repo *SQLiteRepository) {
				createTestSnippet(t, repo, "cmd1", "desc1", nil)
				createTestSnippet(t, repo, "cmd2", "desc2", nil)
			},
			filter: domain.ListFilter{},
			check: func(t *testing.T, snippets []domain.Snippet) {
				assert.Len(t, snippets, 2)
			},
		},
		{
			name: "filters by tag",
			setup: func(t *testing.T, repo *SQLiteRepository) {
				createTestSnippet(t, repo, "docker ps", "containers", []string{"docker"})
				createTestSnippet(t, repo, "git status", "status", []string{"git"})
				createTestSnippet(t, repo, "docker logs", "logs", []string{"docker"})
			},
			filter: domain.ListFilter{Tag: "docker"},
			check: func(t *testing.T, snippets []domain.Snippet) {
				assert.Len(t, snippets, 2)
				for _, s := range snippets {
					assert.Contains(t, s.Command, "docker")
				}
			},
		},
		{
			name: "respects limit",
			setup: func(t *testing.T, repo *SQLiteRepository) {
				createTestSnippet(t, repo, "cmd1", "desc1", nil)
				createTestSnippet(t, repo, "cmd2", "desc2", nil)
				createTestSnippet(t, repo, "cmd3", "desc3", nil)
			},
			filter: domain.ListFilter{Limit: 2},
			check: func(t *testing.T, snippets []domain.Snippet) {
				assert.Len(t, snippets, 2)
			},
		},
		{
			name: "respects offset",
			setup: func(t *testing.T, repo *SQLiteRepository) {
				createTestSnippet(t, repo, "cmd1", "desc1", nil)
				createTestSnippet(t, repo, "cmd2", "desc2", nil)
				createTestSnippet(t, repo, "cmd3", "desc3", nil)
			},
			filter: domain.ListFilter{Limit: 10, Offset: 2},
			check: func(t *testing.T, snippets []domain.Snippet) {
				assert.Len(t, snippets, 1)
			},
		},
		{
			name: "defaults limit to 50",
			setup: func(t *testing.T, repo *SQLiteRepository) {
				createTestSnippet(t, repo, "cmd1", "desc1", nil)
			},
			filter: domain.ListFilter{Limit: 0},
			check: func(t *testing.T, snippets []domain.Snippet) {
				// Should not fail — limit defaults to 50
				assert.Len(t, snippets, 1)
			},
		},
		{
			name: "includes tags for each snippet",
			setup: func(t *testing.T, repo *SQLiteRepository) {
				createTestSnippet(t, repo, "cmd", "desc", []string{"tag1", "tag2"})
			},
			filter: domain.ListFilter{},
			check: func(t *testing.T, snippets []domain.Snippet) {
				require.Len(t, snippets, 1)
				assert.ElementsMatch(t, []string{"tag1", "tag2"}, snippets[0].Tags)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := setupTestDB(t)
			tt.setup(t, repo)

			snippets, err := repo.List(context.Background(), tt.filter)
			require.NoError(t, err)
			tt.check(t, snippets)
		})
	}
}

func TestDelete_CascadesSnippetTags(t *testing.T) {
	repo := setupTestDB(t)
	snippet := createTestSnippet(t, repo, "cmd", "desc", []string{"tag1"})

	err := repo.Delete(context.Background(), snippet.ID)
	require.NoError(t, err)

	// Verify snippet_tags junction rows were cascaded
	var count int
	err = repo.db.QueryRow("SELECT COUNT(*) FROM snippet_tags WHERE snippet_id = ?", snippet.ID).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 0, count)
}

func TestMigrate_Idempotent(t *testing.T) {
	repo := setupTestDB(t)
	// Running migrate again should not error (ErrNoChange is swallowed)
	err := repo.Migrate()
	assert.NoError(t, err)
}
