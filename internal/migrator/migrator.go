package migrator

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gorm.io/gorm"
)

// Admin applies all SQL files in migrations/admin/ against the admin database.
// All statements must be idempotent (IF NOT EXISTS).
func Admin(ctx context.Context, db *gorm.DB, dir string) error {
	return apply(ctx, db, dir, "")
}

// Tenant applies all SQL files in migrations/tenant/ against a tenant database.
// If schema is non-empty, search_path is set first (schema-mode tenants).
func Tenant(ctx context.Context, db *gorm.DB, dir, schema string) error {
	return apply(ctx, db, dir, schema)
}

func apply(ctx context.Context, gormDB *gorm.DB, dir, schema string) error {
	sqlDB, err := gormDB.DB()
	if err != nil {
		return fmt.Errorf("get sql.DB: %w", err)
	}

	if schema != "" {
		if _, err := sqlDB.ExecContext(ctx, "SET search_path TO "+schema); err != nil {
			return fmt.Errorf("set search_path=%q: %w", schema, err)
		}
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("read migrations dir %q: %w", dir, err)
	}

	var files []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") {
			files = append(files, filepath.Join(dir, e.Name()))
		}
	}
	sort.Strings(files) // lexicographic = chronological (YYYYMMDD prefix)

	for _, f := range files {
		content, err := os.ReadFile(f)
		if err != nil {
			return fmt.Errorf("read %s: %w", f, err)
		}
		if _, err := sqlDB.ExecContext(ctx, string(content)); err != nil {
			return fmt.Errorf("exec %s: %w", filepath.Base(f), err)
		}
		fmt.Printf("Applied: %s\n", filepath.Base(f))
	}
	return nil
}
