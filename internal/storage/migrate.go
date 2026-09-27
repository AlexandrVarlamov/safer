package storage

import (
	"context"
	"fmt"
	"io/fs"
	"sort"

	"github.com/aleksandrvarlamov/safer/migrations"
)

// Migrate применяет все ещё не применённые *.sql из каталога migrations.
// Безопасна при повторных запусках: применённые версии отслеживаются
// в таблице schema_migrations.
func (d *D1) Migrate(ctx context.Context) error {
	if _, err := d.Query(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (
		version    TEXT PRIMARY KEY,
		applied_at TEXT NOT NULL DEFAULT (datetime('now'))
	)`); err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}

	applied := make(map[string]bool)
	rows, err := d.Query(ctx, `SELECT version FROM schema_migrations`)
	if err != nil {
		return fmt.Errorf("select applied versions: %w", err)
	}
	for _, row := range rows {
		if v, ok := row["version"].(string); ok {
			applied[v] = true
		}
	}

	files, err := fs.Glob(migrations.FS, "*.sql")
	if err != nil {
		return fmt.Errorf("list migrations: %w", err)
	}
	sort.Strings(files) // порядок даёт числовой префикс имени: 0001, 0002, ...

	for _, name := range files {
		if applied[name] {
			continue
		}

		body, err := migrations.FS.ReadFile(name)
		if err != nil {
			return fmt.Errorf("read %s: %w", name, err)
		}
		if _, err := d.Query(ctx, string(body)); err != nil {
			return fmt.Errorf("apply %s: %w", name, err)
		}
		if _, err := d.Query(ctx,
			`INSERT INTO schema_migrations (version) VALUES (?)`, name); err != nil {
			return fmt.Errorf("record %s: %w", name, err)
		}
	}

	return nil
}
