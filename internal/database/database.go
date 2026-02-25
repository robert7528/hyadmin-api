package database

import (
	"fmt"

	"github.com/hysp/hyadmin-api/internal/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Connect opens the admin database connection.
// Schema migrations are managed by Atlas — use cmd/migrate, not AutoMigrate.
func Connect(cfg *config.Config) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(cfg.Database.DSN), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	return db, nil
}
