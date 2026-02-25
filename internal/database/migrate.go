package database

import (
	"context"
	"fmt"

	"ariga.io/atlas/sql/migrate"
	atlas_postgres "ariga.io/atlas/sql/postgres"
	"gorm.io/gorm"
)

// MigrateAdmin applies all pending Atlas SQL migrations from migrations/admin/
// against the admin database.
func MigrateAdmin(ctx context.Context, db *gorm.DB, dir string) error {
	return applyMigrations(ctx, db, dir, "")
}

// MigrateTenant applies all pending Atlas SQL migrations from migrations/tenant/
// against a tenant's database. If schema is non-empty, search_path is set first
// (for schema-mode tenants sharing one PostgreSQL instance).
func MigrateTenant(ctx context.Context, db *gorm.DB, dir, schema string) error {
	return applyMigrations(ctx, db, dir, schema)
}

func applyMigrations(ctx context.Context, gormDB *gorm.DB, dir, schema string) error {
	sqlDB, err := gormDB.DB()
	if err != nil {
		return fmt.Errorf("get sql.DB: %w", err)
	}

	if schema != "" {
		if _, err := sqlDB.ExecContext(ctx, "SET search_path TO "+schema); err != nil {
			return fmt.Errorf("set search_path=%q: %w", schema, err)
		}
	}

	driver, err := atlas_postgres.Open(sqlDB)
	if err != nil {
		return fmt.Errorf("open atlas driver: %w", err)
	}

	localDir, err := migrate.NewLocalDir(dir)
	if err != nil {
		return fmt.Errorf("open migrations dir %q: %w", dir, err)
	}

	m, err := migrate.NewMigrator(driver, localDir)
	if err != nil {
		return fmt.Errorf("create migrator: %w", err)
	}

	if err := m.ApplyContext(ctx); err != nil {
		return fmt.Errorf("apply migrations: %w", err)
	}
	return nil
}
