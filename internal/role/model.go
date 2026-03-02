package role

import (
	"time"

	"gorm.io/gorm"
)

func (Role) TableName() string    { return "hyadmin_roles" }
func (UserRole) TableName() string { return "hyadmin_user_roles" }

type Role struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	TenantCode  string         `gorm:"index;not null" json:"tenant_code"`
	Name        string         `gorm:"not null" json:"name"`
	Description string         `json:"description"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// UserRole is managed by Casbin (g policy), kept only for GORM auto-migrate reference.
// Actual assignments live in casbin_rule table.
type UserRole struct {
	UserID uint `gorm:"primaryKey"`
	RoleID uint `gorm:"primaryKey"`
}

type CreateRoleRequest struct {
	TenantCode  string `json:"tenant_code" binding:"required"`
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

type UpdateRoleRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type AssignPermissionsRequest struct {
	PermissionIDs []uint `json:"permission_ids" binding:"required"`
}

type AssignUsersRequest struct {
	UserIDs []uint `json:"user_ids" binding:"required"`
}
