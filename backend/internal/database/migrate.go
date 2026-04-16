package database

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"sort"
	"strings"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

func RunMigrations(ctx context.Context, db *sql.DB) error {
	entries, err := fs.ReadDir(migrationsFS, "migrations")
	if err != nil {
		return fmt.Errorf("read migrations: %w", err)
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}

		path := "migrations/" + entry.Name()
		sqlBytes, err := fs.ReadFile(migrationsFS, path)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", entry.Name(), err)
		}

		script := strings.TrimSpace(string(sqlBytes))
		if script == "" {
			continue
		}

		if _, err := db.ExecContext(ctx, script); err != nil {
			return fmt.Errorf("apply migration %s: %w", entry.Name(), err)
		}
	}

	return nil
}
