package tenant

import (
	"time"

	"gorm.io/gorm"
)

type Tenant struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Code        string         `gorm:"uniqueIndex;not null" json:"code"`
	Name        string         `gorm:"not null" json:"name"`
	Enabled     bool           `gorm:"default:true" json:"enabled"`
	// InfraType and InfraConfig are reserved for hyinfra; hyadmin stores but does not parse them.
	InfraType   string         `gorm:"default:'podman'" json:"infra_type"`   // podman|k8s-namespace|k8s-cluster|k8s-multi
	InfraConfig string         `gorm:"type:text" json:"infra_config"`          // JSONB payload for hyinfra
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}
