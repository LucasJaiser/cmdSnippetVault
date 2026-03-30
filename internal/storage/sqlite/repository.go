package sqlite

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"lucasjaiser/goSnipperVault/internal/domain"
	"time"

	"github.com/golang-migrate/migrate/v4"

	_ "github.com/golang-migrate/migrate/v4/database/sqlite"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

// SQLiteRepository implements domain.SnippetRepository using SQLite.
type SQLiteRepository struct {
	db       *sql.DB
	dsn      string
	file_dsn string
}

//go:embed migrations/*.sql
var migrations embed.FS

func (r *SQLiteRepository) Migrate() error {
	source, err := iofs.New(migrations, "migrations")
	if err != nil {
		return fmt.Errorf("could not read migrations: %w", err)
	}

	migrations_instance, err := migrate.NewWithSourceInstance("iofs", source, r.dsn)
	if err != nil {
		return fmt.Errorf("could not create migrations instance: %w", err)
	}

	err = migrations_instance.Up()
	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("could not migrate up: %w", err)
	}

	return nil
}

func (r *SQLiteRepository) Close() error {
	err := r.db.Close()
	if err != nil {
		return fmt.Errorf("could not close database: %w", err)
	}

	return nil
}

func (r *SQLiteRepository) Create(ctx context.Context, snippet *domain.Snippet) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("could not create transaction: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck // no-op after commit

	err = r.createWithTX(ctx, snippet, tx)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("could not commit statements %w", err)
	}

	return nil
}

func (r *SQLiteRepository) GetByID(ctx context.Context, id int64) (*domain.Snippet, error) {
	row := r.db.QueryRowContext(ctx, "SELECT id, command, description, created_at, updated_at, use_count FROM snippets WHERE id = ?", id)

	var snippet domain.Snippet
	var createdAt, updatedAt string
	err := row.Scan(&snippet.ID, &snippet.Command, &snippet.Description, &createdAt, &updatedAt, &snippet.UseCount)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("could not get snippet: %w", err)
	}

	snippet.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	snippet.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)

	snippet.Tags, err = r.getTagsForSnippet(ctx, snippet.ID)
	if err != nil {
		return nil, fmt.Errorf("could not get tags for snippet: %w", err)
	}

	return &snippet, nil
}

func (r *SQLiteRepository) List(ctx context.Context, filter domain.ListFilter) ([]*domain.Snippet, error) {
	query := `SELECT s.id, s.command, s.description, s.use_count, s.created_at, s.updated_at FROM snippets s`
	var args []any

	if filter.Tag != "" {
		query += ` JOIN snippet_tags st ON st.snippet_id = s.id JOIN tags t ON t.id = st.tag_id WHERE t.name = ?`
		args = append(args, filter.Tag)
	}

	query += ` ORDER BY s.updated_at DESC`

	limit := filter.Limit
	if limit <= 0 {
		limit = 50
	}
	query += ` LIMIT ?`
	args = append(args, limit)

	if filter.Offset > 0 {
		query += ` OFFSET ?`
		args = append(args, filter.Offset)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("could not Query snipppets: %w", err)
	}
	defer rows.Close()

	var snippets []*domain.Snippet
	for rows.Next() {
		snippet := &domain.Snippet{}
		var createdAt, updatedAt string
		if err := rows.Scan(&snippet.ID, &snippet.Command, &snippet.Description, &snippet.UseCount, &createdAt, &updatedAt); err != nil {
			return nil, fmt.Errorf("could not scan snippet row: %w", err)
		}
		snippet.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		snippet.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)
		snippets = append(snippets, snippet)
	}

	for i := range snippets {
		tags, err := r.getTagsForSnippet(ctx, snippets[i].ID)
		if err != nil {
			return nil, fmt.Errorf("could get tags for snippet: %w", err)
		}
		snippets[i].Tags = tags
	}

	return snippets, nil
}

func (r *SQLiteRepository) Update(ctx context.Context, snippet *domain.Snippet) error {

	if snippet.ID == 0 {
		return r.Create(ctx, snippet)
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("could not create transaction: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck // no-op after commit

	//Check if Snippets exits in Database
	_, err = r.GetByID(ctx, snippet.ID)
	if err != nil && errors.Is(err, domain.ErrNotFound) {
		return r.Create(ctx, snippet)
	}

	_, err = tx.ExecContext(ctx, "UPDATE snippets SET command = ?, description = ?, updated_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now'), use_count = ? WHERE id = ?", snippet.Command, snippet.Description, snippet.UseCount, snippet.ID)
	if err != nil {
		return fmt.Errorf("could not update snippet: %w", err)
	}
	_, err = tx.ExecContext(ctx, "DELETE FROM snippet_tags WHERE snippet_id = ?", snippet.ID)
	if err != nil {
		return fmt.Errorf("could not delete snippet: %w", err)
	}
	for _, name := range snippet.Tags {

		tag, err := r.createOrGetTag(ctx, name, tx)
		if err != nil {
			return fmt.Errorf("could not create or get tag: %w", err)
		}

		err = r.linkTag(ctx, int(tag.ID), int(snippet.ID), tx)
		if err != nil {
			return fmt.Errorf("could not link tag to snippet: %w", err)
		}

	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("could not commit statemants %w", err)
	}
	return nil
}

func (r *SQLiteRepository) getTagsForSnippet(ctx context.Context, id int64) ([]string, error) {

	rows, err := r.db.QueryContext(ctx, "SELECT t.name FROM tags t JOIN snippet_tags st ON st.tag_id = t.id WHERE st.snippet_id = ?", id)
	if err != nil {
		return nil, fmt.Errorf("could not get corresponding tags: %w", err)
	}

	defer rows.Close()
	var tags []string
	for rows.Next() {
		var tag string
		if err := rows.Scan(&tag); err != nil {
			return nil, fmt.Errorf("could not assign tag to variable: %w", err)
		}
		tags = append(tags, tag)
	}
	return tags, nil
}

func (r *SQLiteRepository) Delete(ctx context.Context, id int64) error {

	_, err := r.db.ExecContext(ctx, "DELETE FROM snippets WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("could not delete snippet: %w", err)
	}

	return nil
}

func (r *SQLiteRepository) Search(ctx context.Context, query string) ([]*domain.Snippet, error) {
	if query == "" {
		return []*domain.Snippet{}, nil
	}

	rows, err := r.db.QueryContext(ctx, `
    SELECT DISTINCT s.id, s.command, s.description, s.use_count, s.created_at, s.updated_at,
        CASE
            WHEN s.command = ? THEN 100
            WHEN s.command LIKE ? COLLATE NOCASE THEN 80
            WHEN s.command LIKE ? COLLATE NOCASE THEN 60
            WHEN s.description LIKE ? COLLATE NOCASE THEN 40
            WHEN t.name LIKE ? COLLATE NOCASE THEN 20
            ELSE 0
        END AS relevance
    FROM snippets s
    LEFT JOIN snippet_tags st ON st.snippet_id = s.id
    LEFT JOIN tags t ON t.id = st.tag_id
    WHERE s.command LIKE ? COLLATE NOCASE
       OR s.description LIKE ? COLLATE NOCASE
       OR t.name LIKE ? COLLATE NOCASE
    ORDER BY relevance DESC, s.use_count DESC
`, query,
		query+"%",
		"%"+query+"%",
		"%"+query+"%",
		"%"+query+"%",
		"%"+query+"%",
		"%"+query+"%",
		"%"+query+"%",
	)
	if err != nil {
		return nil, fmt.Errorf("sqlite: search: %w", err)
	}
	defer rows.Close()

	var snippets []*domain.Snippet
	for rows.Next() {
		s := &domain.Snippet{}
		var createdAt, updatedAt string
		var relevance int
		if err := rows.Scan(&s.ID, &s.Command, &s.Description, &s.UseCount, &createdAt, &updatedAt, &relevance); err != nil {
			return nil, fmt.Errorf("sqlite: scan search result: %w", err)
		}
		s.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		s.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)
		snippets = append(snippets, s)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("sqlite: search rows: %w", err)
	}

	// load tags for each result
	for i := range snippets {
		tags, err := r.getTagsForSnippet(ctx, snippets[i].ID)
		if err != nil {
			return nil, err
		}
		snippets[i].Tags = tags
	}

	return snippets, nil
}

func (r *SQLiteRepository) ListTags(ctx context.Context) ([]*domain.TagWithCount, error) {
	rows, err := r.db.QueryContext(ctx, `
                SELECT t.name, COUNT(st.snippet_id)
                FROM tags t
                LEFT JOIN snippet_tags st ON st.tag_id = t.id
                GROUP BY t.id, t.name
                ORDER BY t.name`)
	if err != nil {
		return nil, fmt.Errorf("sqlite: list tags: %w", err)
	}
	defer rows.Close()

	var tags []*domain.TagWithCount
	for rows.Next() {
		tc := &domain.TagWithCount{}
		if err := rows.Scan(&tc.Name, &tc.Count); err != nil {
			return nil, fmt.Errorf("sqlite: scan tag: %w", err)
		}
		tags = append(tags, tc)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("sqlite: list tags rows: %w", err)
	}
	return tags, nil
}

func (r *SQLiteRepository) createOrGetTag(ctx context.Context, name string, tx *sql.Tx) (*domain.Tag, error) {

	var tag domain.Tag
	err := tx.QueryRowContext(ctx, "SELECT id, name FROM tags WHERE name = ?", name).Scan(&tag.ID, &tag.Name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			_, err := tx.ExecContext(ctx, "INSERT INTO tags (name) VALUES(?) ON CONFLICT (name) DO NOTHING", name)
			if err != nil {
				return nil, fmt.Errorf("could not insert tag: %w", err)
			}

			var id int64
			err = tx.QueryRowContext(ctx, "SELECT id FROM tags WHERE name = ?", name).Scan(&id)
			if err != nil {
				return nil, fmt.Errorf("could not get last inserted id from tag: %w", err)
			}

			tag.ID = id
			tag.Name = name

			return &tag, nil

		}
		return nil, fmt.Errorf("could not get tag: %w", err)

	}

	return &tag, nil
}

func (r *SQLiteRepository) linkTag(ctx context.Context, tag_id int, snippet_id int, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx, "INSERT INTO snippet_tags (tag_id, snippet_id) VALUES (?, ?)", tag_id, snippet_id)
	if err != nil {
		return fmt.Errorf("could not link tag to snippet: %w", err)
	}

	return nil
}

func (r *SQLiteRepository) unlinkTag(ctx context.Context, tag_id int, snippet_id int, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx, "DELETE FROM snippet_tags WHERE tag_id = ? AND snippet_id = ?", tag_id, snippet_id)

	if err != nil {
		return fmt.Errorf("could not unlink tag to snippet: %w", err)
	}

	return nil
}

func (r *SQLiteRepository) CreateBatch(ctx context.Context, snippets []*domain.Snippet) (*domain.ImportStatistics, error) {
	var duplicates int
	var created int

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("could not create transaction: %w", err)
	}

	defer tx.Rollback() //nolint:errcheck // no-op after commit

	for _, snippet := range snippets {

		duplicate, err := r.searchForDuplicate(ctx, snippet.Command, tx)
		if err != nil {
			return nil, err
		}

		if duplicate {
			duplicates += 1
			continue
		}

		// Create Snippet
		err = r.createWithTX(ctx, snippet, tx)
		if err != nil {
			return nil, err
		}

		created += 1
	}

	err = tx.Commit()
	if err != nil {
		return nil, fmt.Errorf("could not commit batch create: %w", err)
	}

	return &domain.ImportStatistics{
		Created:    created,
		Duplicates: duplicates,
	}, nil

}

func (r *SQLiteRepository) searchForDuplicate(ctx context.Context, command string, tx *sql.Tx) (bool, error) {

	result := tx.QueryRowContext(ctx, "SELECT COUNT(*) FROM snippets WHERE command = ?", command)

	var entries int
	err := result.Scan(&entries)
	if err != nil {
		return false, fmt.Errorf("could not scan entries: %w", err)
	}

	if entries > 0 {
		return true, nil
	}

	return false, nil
}

func (r *SQLiteRepository) createWithTX(ctx context.Context, snippet *domain.Snippet, tx *sql.Tx) error {

	snippetResult, err := tx.ExecContext(ctx, "INSERT INTO snippets (`command`, `description`) VALUES (?, ?)", snippet.Command, snippet.Description)

	if err != nil {
		return fmt.Errorf("could not insert command: %w", err)
	}

	snippetID, err := snippetResult.LastInsertId()
	if err != nil {
		return fmt.Errorf("could not get last inserted id from command: %w", err)
	}

	for _, name := range snippet.Tags {

		tag, err := r.createOrGetTag(ctx, name, tx)
		if err != nil {
			return fmt.Errorf("could not create or get tag: %w", err)
		}

		err = r.linkTag(ctx, int(tag.ID), int(snippetID), tx)
		if err != nil {
			return fmt.Errorf("could not link tag to snippet: %w", err)
		}

	}

	snippet.ID = snippetID

	return nil

}
