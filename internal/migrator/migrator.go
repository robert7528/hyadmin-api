package migrator

import (
	"context"
	"fmt"

	atlas_postgres "ariga.io/atlas/sql/postgres"
	"ariga.io/atlas/sql/migrate"
	"gorm.io/gorm"
)

// Admin applies all pending Atlas SQL migrations from migrations/admin/
// against the admin database.
func Admin(ctx context.Context, db *gorm.DB, dir string) error {
	return apply(ctx, db, dir, "")
}

// Tenant applies all pending Atlas SQL migrations from migrations/tenant/
// against a tenant's database. If schema is non-empty, search_path is set first
// (for schema-mode tenants sharing one PostgreSQL instance).
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

	driver, err := atlas_postgres.Open(sqlDB)
	if err != nil {
		return fmt.Errorf("open atlas driver: %w", err)
	}

	localDir, err := migrate.NewLocalDir(dir)
	if err != nil {
		return fmt.Errorf("open migrations dir %q: %w", dir, err)
	}

	ex, err := migrate.NewExecutor(driver, localDir, migrate.NopRevisionReadWriter{}, migrate.WithLogger(migrate.NopLogger{}), migrate.WithAllowDirty(true))
	if err != nil {
		return fmt.Errorf("create executor: %w", err)
	}

	if err := ex.ExecuteN(ctx, 0); err != nil {
		return fmt.Errorf("apply migrations: %w", err)
	}
	return nil
}
