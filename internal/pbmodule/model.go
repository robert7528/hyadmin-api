package pbmodule

import (
	"time"

	"gorm.io/gorm"
)

func (PlatformModule) TableName() string { return "hyadmin_modules" }

// PlatformModule represents a top-level navigation module (tab in header).
type PlatformModule struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Name        string         `gorm:"uniqueIndex;not null" json:"name"`
	DisplayName string         `gorm:"not null" json:"display_name"`
	Icon        string         `json:"icon"`
	Route       string         `gorm:"not null" json:"route"`
	URL         string         `json:"url"`
	Description string         `json:"description"`
	SortOrder   int            `gorm:"default:0" json:"sort_order"`
	Enabled     bool           `gorm:"default:true" json:"enabled"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

type CreateModuleRequest struct {
	Name        string `json:"name" binding:"required"`
	DisplayName string `json:"display_name" binding:"required"`
	Icon        string `json:"icon"`
	Route       string `json:"route" binding:"required"`
	URL         string `json:"url"`
	Description string `json:"description"`
	SortOrder   int    `json:"sort_order"`
}

type UpdateModuleRequest struct {
	DisplayName string `json:"display_name"`
	Icon        string `json:"icon"`
	URL         string `json:"url"`
	Description string `json:"description"`
	SortOrder   *int   `json:"sort_order"`
	Enabled     *bool  `json:"enabled"`
}
