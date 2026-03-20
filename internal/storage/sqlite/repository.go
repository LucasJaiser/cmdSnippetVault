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
		return fmt.Errorf("could not read migrations: %s", err.Error())
	}

	migrations_instance, err := migrate.NewWithSourceInstance("iofs", source, r.dsn)
	if err != nil {
		return fmt.Errorf("could not create migrations instance: %s", err.Error())
	}

	err = migrations_instance.Up()
	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("could not migrate up: %s", err.Error())
	}

	return nil
}

func (r *SQLiteRepository) Close() error {
	err := r.db.Close()
	if err != nil {
		return fmt.Errorf("could not close database: %s", err.Error())
	}

	return nil
}

func (r *SQLiteRepository) Create(ctx context.Context, snippet *domain.Snippet) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("could not create transaction: %s", err.Error())
	}
	defer tx.Rollback()

	snippetResult, err := tx.ExecContext(ctx, "INSERT INTO snippets (`command`, `description`) VALUES (?, ?)", snippet.Command, snippet.Description)

	if err != nil {
		return fmt.Errorf("could not insert command: %s", err.Error())
	}

	snippetID, err := snippetResult.LastInsertId()
	if err != nil {
		return fmt.Errorf("could not get last inserted id from command: %s", err.Error())
	}

	for _, tag := range snippet.Tags {

		_, err := tx.ExecContext(ctx, "INSERT INTO tags (name) VALUES (?) ON CONFLICT (name) DO NOTHING", tag)
		if err != nil {
			return fmt.Errorf("could not insert tag: %s", err.Error())
		}

		var tagID int64
		err = tx.QueryRowContext(ctx, "SELECT id FROM tags WHERE name = ?", tag).Scan(&tagID)
		if err != nil {
			return fmt.Errorf("could not get last inserted id from tag: %s", err.Error())
		}

		_, err = tx.ExecContext(ctx, "INSERT INTO snippet_tags (snippet_id, tag_id) VALUES(?, ?)", snippetID, tagID)
		if err != nil {
			return fmt.Errorf("could not insert snippet_tags: %s", err.Error())
		}

	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("could not commit statements %s", err.Error())
	}

	snippet.ID = snippetID
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
		return nil, fmt.Errorf("could not get snippet: %s", err.Error())
	}

	snippet.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	snippet.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)

	snippet.Tags, err = r.getTagsForSnippet(ctx, snippet.ID)
	if err != nil {
		return nil, fmt.Errorf("could not get tags for snippet: %s", err.Error())
	}

	return &snippet, nil
}

func (r *SQLiteRepository) List(ctx context.Context, filter domain.ListFilter) ([]domain.Snippet, error) {
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
		return nil, fmt.Errorf("could not Query snipppets: %s", err.Error())
	}
	defer rows.Close()

	var snippets []domain.Snippet
	for rows.Next() {
		var snippet domain.Snippet
		var createdAt, updatedAt string
		if err := rows.Scan(&snippet.ID, &snippet.Command, &snippet.Description, &snippet.UseCount, &createdAt, &updatedAt); err != nil {
			return nil, fmt.Errorf("could not scan snippet row: %s", err.Error())
		}
		snippet.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		snippet.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)
		snippets = append(snippets, snippet)
	}

	for i := range snippets {
		tags, err := r.getTagsForSnippet(ctx, snippets[i].ID)
		if err != nil {
			return nil, fmt.Errorf("could get tags for snippet: %s", err.Error())
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
		return fmt.Errorf("could not create transaction: %s", err.Error())
	}
	defer tx.Rollback()

	//Check if Snippets exits in Database
	_, err = r.GetByID(ctx, snippet.ID)
	if err != nil && errors.Is(err, domain.ErrNotFound) {
		return r.Create(ctx, snippet)
	}

	_, err = tx.ExecContext(ctx, "UPDATE snippets SET command = ?, description = ?, updated_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now'), use_count = ? WHERE id = ?", snippet.Command, snippet.Description, snippet.UseCount, snippet.ID)
	if err != nil {
		return fmt.Errorf("could not update snippet: %s", err.Error())
	}
	_, err = tx.ExecContext(ctx, "DELETE FROM snippet_tags WHERE snippet_id = ?", snippet.ID)
	if err != nil {
		return fmt.Errorf("could not delete snippet: %s", err.Error())
	}
	for _, tag := range snippet.Tags {

		_, err := tx.ExecContext(ctx, "INSERT INTO tags (`name`) VALUES (?) ON CONFLICT (name) DO NOTHING", tag)
		if err != nil {
			return fmt.Errorf("could not insert tag: %s", err.Error())
		}

		var tagID int64
		err = tx.QueryRowContext(ctx, "SELECT id FROM tags WHERE name = ?", tag).Scan(&tagID)
		if err != nil {
			return fmt.Errorf("could not get last inserted id from tag: %s", err.Error())
		}

		_, err = tx.ExecContext(ctx, "INSERT INTO snippet_tags (snippet_id, tag_id) VALUES(?, ?)", snippet.ID, tagID)
		if err != nil {
			return fmt.Errorf("could not insert snipper_tags: %s", err.Error())
		}
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("could not commit statemants %s", err.Error())
	}
	return nil
}

func (r *SQLiteRepository) getTagsForSnippet(ctx context.Context, id int64) ([]string, error) {

	rows, err := r.db.QueryContext(ctx, "SELECT t.name FROM tags t JOIN snippet_tags st ON st.tag_id = t.id WHERE st.snippet_id = ?", id)
	if err != nil {
		return nil, fmt.Errorf("could not get corresponding tags: %s", err.Error())
	}

	defer rows.Close()
	var tags []string
	for rows.Next() {
		var tag string
		if err := rows.Scan(&tag); err != nil {
			return nil, fmt.Errorf("could not assign tag to variable: %s", err.Error())
		}
		tags = append(tags, tag)
	}
	return tags, nil
}

func (r *SQLiteRepository) Delete(ctx context.Context, id int64) error {

	_, err := r.db.ExecContext(ctx, "DELETE FROM snippets WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("could not delete snippet: %s", err.Error())
	}

	return nil
}

func (r *SQLiteRepository) Search(ctx context.Context, query string) ([]domain.Snippet, error) {
	return nil, nil
}
