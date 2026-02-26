package adminuser

import (
	"time"

	"gorm.io/gorm"
)

// AdminUser stores platform administrators.
// PII fields (display_name, email) are encrypted with Tink before persisting.
type AdminUser struct {
	ID             uint           `gorm:"primaryKey" json:"id"`
	TenantCode     string         `gorm:"index;not null" json:"tenant_code"`
	Username       string         `gorm:"uniqueIndex:uk_tenant_user;not null" json:"username"`
	PasswordHash   string         `json:"-"`                      // bcrypt; empty for third-party logins
	DisplayNameEnc string         `gorm:"column:display_name" json:"-"` // Tink-encrypted
	EmailEnc       string         `gorm:"column:email" json:"-"`        // Tink-encrypted
	Provider       string         `gorm:"default:'local'" json:"provider"` // local|google|...
	ProviderID     string         `json:"provider_id,omitempty"`
	Enabled        bool           `gorm:"default:true" json:"enabled"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

// AdminUserDTO is the decrypted representation returned to callers.
type AdminUserDTO struct {
	ID          uint      `json:"id"`
	TenantCode  string    `json:"tenant_code"`
	Username    string    `json:"username"`
	DisplayName string    `json:"display_name"`
	Email       string    `json:"email"`
	Provider    string    `json:"provider"`
	Enabled     bool      `json:"enabled"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CreateUserRequest struct {
	TenantCode  string `json:"tenant_code" binding:"required"`
	Username    string `json:"username" binding:"required"`
	Password    string `json:"password"`
	DisplayName string `json:"display_name"`
	Email       string `json:"email"`
	Provider    string `json:"provider"`
	ProviderID  string `json:"provider_id"`
}

type UpdateUserRequest struct {
	DisplayName string `json:"display_name"`
	Email       string `json:"email"`
	Enabled     *bool  `json:"enabled"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}
