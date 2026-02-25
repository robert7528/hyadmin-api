package database

import "time"

// TenantDBConfig stores per-tenant database connection settings in the admin DB.
// Mode "database": each tenant uses its own database (different DSNs).
// Mode "schema":   tenants share the same database but use separate PostgreSQL schemas.
type TenantDBConfig struct {
	ID         uint   `gorm:"primaryKey" json:"id"`
	TenantCode string `gorm:"uniqueIndex;not null" json:"tenant_code"`

	// "database" | "schema"
	Mode string `gorm:"not null;default:'database'" json:"mode"`

	// Primary (write) connection DSN.
	// Supports libpq key=value format or postgres:// URL format.
	// TODO: encrypt at rest before storing.
	PrimaryDSN string `gorm:"not null" json:"-"`

	// Replica (read) DSNs — JSON-encoded []string.
	// Empty means no read replicas (all queries go to primary).
	// TODO: encrypt at rest before storing.
	ReplicaDSNs string `gorm:"type:text" json:"-"`

	// PostgreSQL schema name — used only when Mode == "schema".
	// If set, search_path is applied to every connection automatically.
	Schema string `json:"schema,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
