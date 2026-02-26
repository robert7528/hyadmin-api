package auditlog

import "time"

func (AuditLog) TableName() string { return "hyadmin_audit_logs" }

// AuditLog records every write action performed by admin users.
type AuditLog struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	TenantCode string    `gorm:"index;not null" json:"tenant_code"`
	UserID     uint      `gorm:"index" json:"user_id"`
	Username   string    `json:"username"`
	Action     string    `json:"action"`     // CREATE|UPDATE|DELETE|LOGIN|LOGOUT
	Resource   string    `json:"resource"`   // e.g. "modules", "users"
	ResourceID string    `json:"resource_id"`
	Detail     string    `gorm:"type:text" json:"detail"` // JSON diff or description
	IP         string    `json:"ip"`
	UserAgent  string    `json:"user_agent"`
	CreatedAt  time.Time `gorm:"index" json:"created_at"`
}
