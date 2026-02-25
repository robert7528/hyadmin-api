package database

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"
)

// DBManager maintains a pool of per-tenant *gorm.DB connections.
// Connections are created on first use and cached for reuse.
type DBManager struct {
	adminDB   *gorm.DB
	tenantDBs sync.Map // map[tenantCode]*gorm.DB
}

func NewManager(adminDB *gorm.DB) *DBManager {
	return &DBManager{adminDB: adminDB}
}

// GetDB returns the *gorm.DB for the given tenant code.
// On first call, it loads the TenantDBConfig from the admin DB, builds the
// connection (with dbresolver if replicas are configured), and caches it.
func (m *DBManager) GetDB(tenantCode string) (*gorm.DB, error) {
	if db, ok := m.tenantDBs.Load(tenantCode); ok {
		return db.(*gorm.DB), nil
	}

	var cfg TenantDBConfig
	if err := m.adminDB.Where("tenant_code = ?", tenantCode).First(&cfg).Error; err != nil {
		return nil, fmt.Errorf("tenant DB config not found for %q: %w", tenantCode, err)
	}

	db, err := buildTenantDB(&cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to build DB for tenant %q: %w", tenantCode, err)
	}

	// LoadOrStore avoids race: if another goroutine already stored it, use that.
	actual, _ := m.tenantDBs.LoadOrStore(tenantCode, db)
	return actual.(*gorm.DB), nil
}

// InvalidateCache drops the cached connection for a tenant.
// Call this after updating a tenant's TenantDBConfig.
func (m *DBManager) InvalidateCache(tenantCode string) {
	m.tenantDBs.Delete(tenantCode)
}

// buildTenantDB creates a *gorm.DB for a tenant, applying:
//   - search_path for schema mode
//   - dbresolver for read/write separation if replicas are configured
func buildTenantDB(cfg *TenantDBConfig) (*gorm.DB, error) {
	primaryDSN := withSearchPath(cfg.PrimaryDSN, cfg.Mode, cfg.Schema)

	db, err := gorm.Open(postgres.Open(primaryDSN), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("open primary: %w", err)
	}

	replicas, err := parseReplicaDSNs(cfg.ReplicaDSNs, cfg.Mode, cfg.Schema)
	if err != nil {
		return nil, err
	}

	if len(replicas) > 0 {
		dialectors := make([]gorm.Dialector, len(replicas))
		for i, dsn := range replicas {
			dialectors[i] = postgres.Open(dsn)
		}
		if err := db.Use(dbresolver.Register(dbresolver.Config{
			// Sources is already the primary opened above.
			// Replicas handle all read (SELECT) queries.
			Replicas: dialectors,
			Policy:   dbresolver.RandomPolicy{},
		})); err != nil {
			return nil, fmt.Errorf("register dbresolver: %w", err)
		}
	}

	return db, nil
}

// withSearchPath appends search_path to the DSN when mode == "schema".
// Supports both URL (postgres://...) and libpq key=value formats.
func withSearchPath(dsn, mode, schema string) string {
	if mode != "schema" || schema == "" {
		return dsn
	}
	if strings.HasPrefix(dsn, "postgres://") || strings.HasPrefix(dsn, "postgresql://") {
		sep := "?"
		if strings.Contains(dsn, "?") {
			sep = "&"
		}
		return dsn + sep + "search_path=" + schema
	}
	// libpq key=value: append options runtime parameter
	return dsn + " options=-csearch_path=" + schema
}

// parseReplicaDSNs unmarshals the JSON replica DSN list and applies search_path.
func parseReplicaDSNs(raw, mode, schema string) ([]string, error) {
	if raw == "" {
		return nil, nil
	}
	var dsns []string
	if err := json.Unmarshal([]byte(raw), &dsns); err != nil {
		return nil, fmt.Errorf("parse replica DSNs: %w", err)
	}
	for i, dsn := range dsns {
		dsns[i] = withSearchPath(dsn, mode, schema)
	}
	return dsns, nil
}
