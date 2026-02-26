// Package casbinx provides a Casbin enforcer backed by GORM (PostgreSQL).
package casbinx

import (
	"github.com/casbin/casbin/v2"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"gorm.io/gorm"
)

// NewEnforcer creates a Casbin enforcer using the GORM adapter.
// The model is loaded from the given conf file path.
// EnableAutoSave ensures policy changes persist to casbin_rule automatically.
func NewEnforcer(db *gorm.DB, modelPath string) (*casbin.Enforcer, error) {
	adapter, err := gormadapter.NewAdapterByDB(db)
	if err != nil {
		return nil, err
	}
	e, err := casbin.NewEnforcer(modelPath, adapter)
	if err != nil {
		return nil, err
	}
	e.EnableAutoSave(true)
	return e, nil
}
