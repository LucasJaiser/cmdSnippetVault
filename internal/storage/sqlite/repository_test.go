package sqlite

import (
	"context"
	"database/sql"
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

func TestSearch(t *testing.T) {
	tests := []struct {
		name  string
		setup func(t *testing.T, repo *SQLiteRepository)
		query string
		check func(t *testing.T, snippets []domain.Snippet)
	}{
		{
			name:  "returns empty slice for empty query",
			setup: func(t *testing.T, repo *SQLiteRepository) {},
			query: "",
			check: func(t *testing.T, snippets []domain.Snippet) {
				assert.Empty(t, snippets)
			},
		},
		{
			name:  "returns empty slice when no snippets exist",
			setup: func(t *testing.T, repo *SQLiteRepository) {},
			query: "git",
			check: func(t *testing.T, snippets []domain.Snippet) {
				assert.Empty(t, snippets)
			},
		},
		{
			name: "matches exact command",
			setup: func(t *testing.T, repo *SQLiteRepository) {
				createTestSnippet(t, repo, "git status", "check working tree", []string{"git"})
				createTestSnippet(t, repo, "docker ps", "list containers", []string{"docker"})
			},
			query: "git status",
			check: func(t *testing.T, snippets []domain.Snippet) {
				require.Len(t, snippets, 1)
				assert.Equal(t, "git status", snippets[0].Command)
			},
		},
		{
			name: "matches command substring",
			setup: func(t *testing.T, repo *SQLiteRepository) {
				createTestSnippet(t, repo, "git status", "check working tree", nil)
				createTestSnippet(t, repo, "git log --oneline", "short log", nil)
				createTestSnippet(t, repo, "docker ps", "list containers", nil)
			},
			query: "git",
			check: func(t *testing.T, snippets []domain.Snippet) {
				assert.Len(t, snippets, 2)
			},
		},
		{
			name: "matches description",
			setup: func(t *testing.T, repo *SQLiteRepository) {
				createTestSnippet(t, repo, "kubectl get pods", "list running containers", nil)
				createTestSnippet(t, repo, "ls -la", "list files", nil)
			},
			query: "containers",
			check: func(t *testing.T, snippets []domain.Snippet) {
				require.Len(t, snippets, 1)
				assert.Equal(t, "kubectl get pods", snippets[0].Command)
			},
		},
		{
			name: "matches tag name",
			setup: func(t *testing.T, repo *SQLiteRepository) {
				createTestSnippet(t, repo, "helm install", "deploy chart", []string{"kubernetes"})
				createTestSnippet(t, repo, "ls -la", "list files", []string{"filesystem"})
			},
			query: "kubernetes",
			check: func(t *testing.T, snippets []domain.Snippet) {
				require.Len(t, snippets, 1)
				assert.Equal(t, "helm install", snippets[0].Command)
			},
		},
		{
			name: "is case insensitive",
			setup: func(t *testing.T, repo *SQLiteRepository) {
				createTestSnippet(t, repo, "Docker Build .", "build image", []string{"docker"})
			},
			query: "docker",
			check: func(t *testing.T, snippets []domain.Snippet) {
				require.Len(t, snippets, 1)
				assert.Equal(t, "Docker Build .", snippets[0].Command)
			},
		},
		{
			name: "includes tags in results",
			setup: func(t *testing.T, repo *SQLiteRepository) {
				createTestSnippet(t, repo, "git push", "push changes", []string{"git", "remote"})
			},
			query: "push",
			check: func(t *testing.T, snippets []domain.Snippet) {
				require.Len(t, snippets, 1)
				assert.ElementsMatch(t, []string{"git", "remote"}, snippets[0].Tags)
			},
		},
		{
			name: "returns no duplicates when matching command and tag",
			setup: func(t *testing.T, repo *SQLiteRepository) {
				createTestSnippet(t, repo, "docker ps", "list docker containers", []string{"docker"})
			},
			query: "docker",
			check: func(t *testing.T, snippets []domain.Snippet) {
				assert.Len(t, snippets, 1)
			},
		},
		{
			name: "ranks exact command match above substring",
			setup: func(t *testing.T, repo *SQLiteRepository) {
				createTestSnippet(t, repo, "git log --oneline", "short log", nil)
				createTestSnippet(t, repo, "git", "base git command", nil)
			},
			query: "git",
			check: func(t *testing.T, snippets []domain.Snippet) {
				require.Len(t, snippets, 2)
				assert.Equal(t, "git", snippets[0].Command)
			},
		},
		{
			name: "returns no results for unmatched query",
			setup: func(t *testing.T, repo *SQLiteRepository) {
				createTestSnippet(t, repo, "git status", "check status", []string{"git"})
			},
			query: "nonexistent",
			check: func(t *testing.T, snippets []domain.Snippet) {
				assert.Empty(t, snippets)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := setupTestDB(t)
			tt.setup(t, repo)

			snippets, err := repo.Search(context.Background(), tt.query)
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

func beginTestTx(t *testing.T, repo *SQLiteRepository) *sql.Tx {
	t.Helper()
	tx, err := repo.db.BeginTx(context.Background(), nil)
	require.NoError(t, err)
	return tx
}

func createTestTag(t *testing.T, repo *SQLiteRepository, name string) int64 {
	t.Helper()
	tx := beginTestTx(t, repo)
	tag, err := repo.createOrGetTag(context.Background(), name, tx)
	require.NoError(t, err)
	require.NoError(t, tx.Commit())
	return tag.ID
}

func TestListTags(t *testing.T) {
	tests := []struct {
		name  string
		setup func(t *testing.T, repo *SQLiteRepository)
		check func(t *testing.T, tags []domain.TagWithCount)
	}{
		{
			name:  "returns empty slice when no tags exist",
			setup: func(t *testing.T, repo *SQLiteRepository) {},
			check: func(t *testing.T, tags []domain.TagWithCount) {
				assert.Empty(t, tags)
			},
		},
		{
			name: "returns tags with snippet counts",
			setup: func(t *testing.T, repo *SQLiteRepository) {
				createTestSnippet(t, repo, "docker ps", "containers", []string{"docker"})
				createTestSnippet(t, repo, "docker logs", "logs", []string{"docker"})
				createTestSnippet(t, repo, "git status", "status", []string{"git"})
			},
			check: func(t *testing.T, tags []domain.TagWithCount) {
				require.Len(t, tags, 2)
				assert.Equal(t, "docker", tags[0].Name)
				assert.Equal(t, 2, tags[0].Count)
				assert.Equal(t, "git", tags[1].Name)
				assert.Equal(t, 1, tags[1].Count)
			},
		},
		{
			name: "returns tags ordered alphabetically",
			setup: func(t *testing.T, repo *SQLiteRepository) {
				createTestSnippet(t, repo, "cmd1", "desc", []string{"zulu"})
				createTestSnippet(t, repo, "cmd2", "desc", []string{"alpha"})
				createTestSnippet(t, repo, "cmd3", "desc", []string{"mike"})
			},
			check: func(t *testing.T, tags []domain.TagWithCount) {
				require.Len(t, tags, 3)
				assert.Equal(t, "alpha", tags[0].Name)
				assert.Equal(t, "mike", tags[1].Name)
				assert.Equal(t, "zulu", tags[2].Name)
			},
		},
		{
			name: "counts only linked snippets",
			setup: func(t *testing.T, repo *SQLiteRepository) {
				s := createTestSnippet(t, repo, "cmd", "desc", []string{"keep", "remove"})
				s.Tags = []string{"keep"}
				err := repo.Update(context.Background(), s)
				require.NoError(t, err)
			},
			check: func(t *testing.T, tags []domain.TagWithCount) {
				for _, tag := range tags {
					if tag.Name == "keep" {
						assert.Equal(t, 1, tag.Count)
					}
					if tag.Name == "remove" {
						assert.Equal(t, 0, tag.Count)
					}
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := setupTestDB(t)
			tt.setup(t, repo)

			tags, err := repo.ListTags(context.Background())
			require.NoError(t, err)
			tt.check(t, tags)
		})
	}
}

func TestLinkTag(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(t *testing.T, repo *SQLiteRepository) (int, int)
		wantErr bool
		check   func(t *testing.T, repo *SQLiteRepository, tagID int, snippetID int)
	}{
		{
			name: "links tag to snippet",
			setup: func(t *testing.T, repo *SQLiteRepository) (int, int) {
				snippet := createTestSnippet(t, repo, "cmd", "desc", nil)
				tagID := createTestTag(t, repo, "newtag")
				return int(tagID), int(snippet.ID)
			},
			check: func(t *testing.T, repo *SQLiteRepository, tagID int, snippetID int) {
				tags, err := repo.getTagsForSnippet(context.Background(), int64(snippetID))
				require.NoError(t, err)
				assert.Contains(t, tags, "newtag")
			},
		},
		{
			name: "duplicate link returns error",
			setup: func(t *testing.T, repo *SQLiteRepository) (int, int) {
				snippet := createTestSnippet(t, repo, "cmd", "desc", []string{"existing"})
				tagID := createTestTag(t, repo, "existing")
				return int(tagID), int(snippet.ID)
			},
			wantErr: true,
		},
		{
			name: "links multiple tags to same snippet",
			setup: func(t *testing.T, repo *SQLiteRepository) (int, int) {
				snippet := createTestSnippet(t, repo, "cmd", "desc", []string{"first"})
				tagID := createTestTag(t, repo, "second")
				return int(tagID), int(snippet.ID)
			},
			check: func(t *testing.T, repo *SQLiteRepository, tagID int, snippetID int) {
				tags, err := repo.getTagsForSnippet(context.Background(), int64(snippetID))
				require.NoError(t, err)
				assert.ElementsMatch(t, []string{"first", "second"}, tags)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := setupTestDB(t)
			tagID, snippetID := tt.setup(t, repo)

			tx := beginTestTx(t, repo)
			err := repo.linkTag(context.Background(), tagID, snippetID, tx)
			if tt.wantErr {
				assert.Error(t, err)
				tx.Rollback() //nolint:errcheck
				return
			}
			require.NoError(t, err)
			require.NoError(t, tx.Commit())
			if tt.check != nil {
				tt.check(t, repo, tagID, snippetID)
			}
		})
	}
}

func TestUnlinkTag(t *testing.T) {
	tests := []struct {
		name  string
		setup func(t *testing.T, repo *SQLiteRepository) (int, int)
		check func(t *testing.T, repo *SQLiteRepository, tagID int, snippetID int)
	}{
		{
			name: "removes link between tag and snippet",
			setup: func(t *testing.T, repo *SQLiteRepository) (int, int) {
				snippet := createTestSnippet(t, repo, "cmd", "desc", []string{"removeme"})
				tagID := createTestTag(t, repo, "removeme")
				return int(tagID), int(snippet.ID)
			},
			check: func(t *testing.T, repo *SQLiteRepository, tagID int, snippetID int) {
				tags, err := repo.getTagsForSnippet(context.Background(), int64(snippetID))
				require.NoError(t, err)
				assert.Empty(t, tags)
			},
		},
		{
			name: "no error when unlinking nonexistent link",
			setup: func(t *testing.T, repo *SQLiteRepository) (int, int) {
				return 9999, 9999
			},
			check: func(t *testing.T, repo *SQLiteRepository, tagID int, snippetID int) {
				// Should succeed silently — DELETE affects zero rows
			},
		},
		{
			name: "only removes specified tag, keeps others",
			setup: func(t *testing.T, repo *SQLiteRepository) (int, int) {
				snippet := createTestSnippet(t, repo, "cmd", "desc", []string{"keep", "remove"})
				tagID := createTestTag(t, repo, "remove")
				return int(tagID), int(snippet.ID)
			},
			check: func(t *testing.T, repo *SQLiteRepository, tagID int, snippetID int) {
				tags, err := repo.getTagsForSnippet(context.Background(), int64(snippetID))
				require.NoError(t, err)
				assert.Equal(t, []string{"keep"}, tags)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := setupTestDB(t)
			tagID, snippetID := tt.setup(t, repo)

			tx := beginTestTx(t, repo)
			err := repo.unlinkTag(context.Background(), tagID, snippetID, tx)
			require.NoError(t, err)
			require.NoError(t, tx.Commit())
			if tt.check != nil {
				tt.check(t, repo, tagID, snippetID)
			}
		})
	}
}

func TestMigrate_Idempotent(t *testing.T) {
	repo := setupTestDB(t)
	// Running migrate again should not error (ErrNoChange is swallowed)
	err := repo.Migrate()
	assert.NoError(t, err)
}

func TestNew_InvalidPath(t *testing.T) {
	// Path with null byte is invalid for SQLite
	_, err := New("/nonexistent/deeply/nested/\x00invalid/path/db.sqlite")
	assert.Error(t, err)
}

func TestCreate_CancelledContext(t *testing.T) {
	repo := setupTestDB(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	snippet := &domain.Snippet{Command: "test", Description: "test"}
	err := repo.Create(ctx, snippet)
	assert.Error(t, err)
}

func TestGetByID_CancelledContext(t *testing.T) {
	repo := setupTestDB(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := repo.GetByID(ctx, 1)
	assert.Error(t, err)
}

func TestList_CancelledContext(t *testing.T) {
	repo := setupTestDB(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := repo.List(ctx, domain.ListFilter{})
	assert.Error(t, err)
}

func TestUpdate_CancelledContext(t *testing.T) {
	repo := setupTestDB(t)
	snippet := createTestSnippet(t, repo, "cmd", "desc", nil)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	snippet.Command = "updated"
	err := repo.Update(ctx, snippet)
	assert.Error(t, err)
}

func TestDelete_CancelledContext(t *testing.T) {
	repo := setupTestDB(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := repo.Delete(ctx, 1)
	assert.Error(t, err)
}

func TestSearch_CancelledContext(t *testing.T) {
	repo := setupTestDB(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := repo.Search(ctx, "test")
	assert.Error(t, err)
}

func TestListTags_CancelledContext(t *testing.T) {
	repo := setupTestDB(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := repo.ListTags(ctx)
	assert.Error(t, err)
}

func TestClose(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "close_test.db")
	repo, err := New(dbPath)
	require.NoError(t, err)

	err = repo.Close()
	assert.NoError(t, err)

	// Operations after close should fail
	_, err = repo.GetByID(context.Background(), 1)
	assert.Error(t, err)
}

func TestCreate_ClosedDB(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "closed_test.db")
	repo, err := New(dbPath)
	require.NoError(t, err)
	repo.Close()

	snippet := &domain.Snippet{Command: "test", Description: "test", Tags: []string{"tag"}}
	err = repo.Create(context.Background(), snippet)
	assert.Error(t, err)
}

func TestUpdate_ClosedDB(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "closed_test.db")
	repo, err := New(dbPath)
	require.NoError(t, err)

	snippet := &domain.Snippet{Command: "cmd", Description: "desc"}
	err = repo.Create(context.Background(), snippet)
	require.NoError(t, err)

	repo.Close()

	snippet.Command = "updated"
	err = repo.Update(context.Background(), snippet)
	assert.Error(t, err)
}

func TestList_ClosedDB(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "closed_test.db")
	repo, err := New(dbPath)
	require.NoError(t, err)
	repo.Close()

	_, err = repo.List(context.Background(), domain.ListFilter{})
	assert.Error(t, err)
}

func TestSearch_ClosedDB(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "closed_test.db")
	repo, err := New(dbPath)
	require.NoError(t, err)
	repo.Close()

	_, err = repo.Search(context.Background(), "test")
	assert.Error(t, err)
}

func TestListTags_ClosedDB(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "closed_test.db")
	repo, err := New(dbPath)
	require.NoError(t, err)
	repo.Close()

	_, err = repo.ListTags(context.Background())
	assert.Error(t, err)
}
