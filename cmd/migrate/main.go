package main

import (
	"fmt"

	"github.com/hysp/hyadmin-api/internal/config"
	"github.com/hysp/hyadmin-api/internal/database"
)

func main() {
	cfg := config.Load()
	db, err := database.Connect(cfg)
	if err != nil {
		panic(fmt.Errorf("failed to connect to database: %w", err))
	}
	if err := database.Migrate(db); err != nil {
		panic(fmt.Errorf("migration failed: %w", err))
	}
	fmt.Println("Migration completed successfully")
}
