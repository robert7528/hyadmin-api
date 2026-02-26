package feature

import (
	"time"

	"gorm.io/gorm"
)

// Feature is a sidebar menu item belonging to a PlatformModule.
type Feature struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	ModuleID  uint           `gorm:"index;not null" json:"module_id"`
	Name      string         `gorm:"not null" json:"name"`
	DisplayName string       `gorm:"not null" json:"display_name"`
	Icon      string         `json:"icon"`
	Path      string         `gorm:"not null" json:"path"` // URL path appended to module route
	SortOrder int            `gorm:"default:0" json:"sort_order"`
	Enabled   bool           `gorm:"default:true" json:"enabled"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type CreateFeatureRequest struct {
	ModuleID    uint   `json:"module_id" binding:"required"`
	Name        string `json:"name" binding:"required"`
	DisplayName string `json:"display_name" binding:"required"`
	Icon        string `json:"icon"`
	Path        string `json:"path" binding:"required"`
	SortOrder   int    `json:"sort_order"`
}

type UpdateFeatureRequest struct {
	DisplayName string `json:"display_name"`
	Icon        string `json:"icon"`
	Path        string `json:"path"`
	SortOrder   *int   `json:"sort_order"`
	Enabled     *bool  `json:"enabled"`
}
