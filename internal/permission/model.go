package permission

import (
	"time"

	"gorm.io/gorm"
)

// Permission is a fine-grained access control point belonging to a Feature.
// Code format: {module}.{feature}.{action}  e.g. "users.list.delete"
type Permission struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	FeatureID   uint           `gorm:"index;not null" json:"feature_id"`
	Code        string         `gorm:"uniqueIndex;not null" json:"code"`
	Name        string         `gorm:"not null" json:"name"`
	Description string         `json:"description"`
	Type        string         `gorm:"not null;default:'button'" json:"type"` // menu|button|api
	SortOrder   int            `gorm:"default:0" json:"sort_order"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// RolePermission links roles to permissions (managed by Casbin p policies).
// Kept for GORM schema reference only.
type RolePermission struct {
	RoleID       uint `gorm:"primaryKey"`
	PermissionID uint `gorm:"primaryKey"`
}

type CreatePermissionRequest struct {
	FeatureID   uint   `json:"feature_id" binding:"required"`
	Code        string `json:"code" binding:"required"`
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Type        string `json:"type"` // menu|button|api
	SortOrder   int    `json:"sort_order"`
}

// BatchCreateRequest generates permissions from standard suffixes.
type BatchCreateRequest struct {
	FeatureID uint     `json:"feature_id" binding:"required"`
	CodePrefix string  `json:"code_prefix" binding:"required"` // e.g. "users.list"
	Suffixes   []string `json:"suffixes" binding:"required"`   // e.g. ["view","create","update","delete"]
}

type UpdatePermissionRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Type        string `json:"type"`
	SortOrder   *int   `json:"sort_order"`
}

// DefaultSuffixes provides common permission templates.
var DefaultSuffixes = map[string]string{
	"view":   "頁面存取",
	"create": "新增",
	"update": "編輯",
	"delete": "刪除",
	"export": "匯出",
}
