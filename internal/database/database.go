package database

import (
	"fmt"

	"github.com/hysp/hyadmin-api/internal/config"
	"github.com/hysp/hyadmin-api/internal/tenant"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Connect(cfg *config.Config) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(cfg.Database.DSN), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	return db, nil
}

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&tenant.Tenant{},
		&TenantDBConfig{},
	)
}
