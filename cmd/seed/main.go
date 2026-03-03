// cmd/seed seeds initial data into the admin database.
// It is idempotent: re-running never overwrites existing rows.
//
// Usage:
//
//	./hyadmin-seed
//
// Environment: same as hyadmin-api (DATABASE_DSN, TINK_KEYSET, etc.)
package main

import (
	"fmt"
	"log"
	"time"

	"github.com/hysp/hyadmin-api/internal/adminuser"
	"github.com/hysp/hyadmin-api/internal/config"
	"github.com/hysp/hyadmin-api/internal/crypto"
	"github.com/hysp/hyadmin-api/internal/database"
	"github.com/hysp/hyadmin-api/internal/pbmodule"
	"github.com/hysp/hyadmin-api/internal/role"
	"github.com/hysp/hyadmin-api/internal/tenant"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func main() {
	cfg := config.Load()

	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("seed: connect DB: %v", err)
	}

	enc, err := crypto.New(cfg.Tink.Keyset)
	if err != nil {
		log.Fatalf("seed: init encryptor: %v", err)
	}

	if err := run(db, enc); err != nil {
		log.Fatalf("seed: %v", err)
	}
	fmt.Println("=== [seed] Completed successfully ===")
}

func run(db *gorm.DB, enc crypto.Encryptor) error {
	// ── 1. System tenant ──────────────────────────────────────
	sysTenant := tenant.Tenant{
		Code:      "system",
		Name:      "HySP System",
		Enabled:   true,
		InfraType: "podman",
	}
	if err := db.Where(tenant.Tenant{Code: "system"}).
		FirstOrCreate(&sysTenant).Error; err != nil {
		return fmt.Errorf("upsert system tenant: %w", err)
	}
	fmt.Printf("  tenant: id=%d code=%s\n", sysTenant.ID, sysTenant.Code)

	// ── 2. Admin user ──────────────────────────────────────────
	displayNameEnc, err := enc.Encrypt("系統管理員")
	if err != nil {
		return fmt.Errorf("encrypt display_name: %w", err)
	}
	hash, err := bcrypt.GenerateFromPassword([]byte("Admin@123456"), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("bcrypt: %w", err)
	}

	adminUser := adminuser.AdminUser{
		TenantCode:     "system",
		Username:       "admin",
		PasswordHash:   string(hash),
		DisplayNameEnc: displayNameEnc,
		Provider:       "local",
		Enabled:        true,
	}
	// Only create if not exists; never overwrite password.
	existing := adminuser.AdminUser{}
	res := db.Where("tenant_code = ? AND username = ?", "system", "admin").
		First(&existing)
	if res.Error == gorm.ErrRecordNotFound {
		if err := db.Create(&adminUser).Error; err != nil {
			return fmt.Errorf("create admin user: %w", err)
		}
		fmt.Printf("  user: id=%d username=%s\n", adminUser.ID, adminUser.Username)
	} else if res.Error != nil {
		return fmt.Errorf("query admin user: %w", res.Error)
	} else {
		adminUser = existing
		fmt.Printf("  user: id=%d username=%s (already exists)\n", adminUser.ID, adminUser.Username)
	}

	// ── 3. Platform modules ────────────────────────────────────
	now := time.Now()
	modules := []pbmodule.PlatformModule{
		{
			Name:        "tenants",
			DisplayName: "租戶管理",
			I18n:        `{"zh-TW":"租戶管理","en":"Tenant Management"}`,
			Route:       "/tenants",
			SortOrder:   1,
			Enabled:     true,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			Name:        "users",
			DisplayName: "使用者管理",
			I18n:        `{"zh-TW":"使用者管理","en":"User Management"}`,
			Route:       "/users",
			SortOrder:   2,
			Enabled:     true,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			Name:        "rbac",
			DisplayName: "權限管理",
			I18n:        `{"zh-TW":"權限管理","en":"Access Control"}`,
			Route:       "/rbac",
			SortOrder:   3,
			Enabled:     true,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			Name:        "audit",
			DisplayName: "稽核日誌",
			I18n:        `{"zh-TW":"稽核日誌","en":"Audit Logs"}`,
			Route:       "/audit",
			SortOrder:   4,
			Enabled:     true,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			Name:        "settings",
			DisplayName: "系統設定",
			I18n:        `{"zh-TW":"系統設定","en":"System Settings"}`,
			Route:       "/settings",
			SortOrder:   5,
			Enabled:     true,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
	}
	for i := range modules {
		m := &modules[i]
		if err := db.Where(pbmodule.PlatformModule{Name: m.Name}).
			FirstOrCreate(m).Error; err != nil {
			return fmt.Errorf("upsert module %q: %w", m.Name, err)
		}
		fmt.Printf("  module: id=%d name=%s\n", m.ID, m.Name)
	}

	// ── 4. Super admin role ────────────────────────────────────
	superRole := role.Role{
		TenantCode:  "system",
		Name:        "super_admin",
		Description: "系統超級管理員，擁有所有權限",
	}
	if err := db.Where(role.Role{TenantCode: "system", Name: "super_admin"}).
		FirstOrCreate(&superRole).Error; err != nil {
		return fmt.Errorf("upsert super_admin role: %w", err)
	}
	fmt.Printf("  role: id=%d name=%s\n", superRole.ID, superRole.Name)

	// ── 5. User → Role assignment ──────────────────────────────
	userRole := role.UserRole{
		UserID: adminUser.ID,
		RoleID: superRole.ID,
	}
	if err := db.Clauses(clause.OnConflict{DoNothing: true}).
		Create(&userRole).Error; err != nil {
		return fmt.Errorf("upsert user_role: %w", err)
	}
	fmt.Printf("  user_role: user_id=%d role_id=%d\n", adminUser.ID, superRole.ID)

	// ── 6. Casbin policies ─────────────────────────────────────
	// g policy: user:N has role:M
	// p policy: role:M can access * with act=access  (super_admin wildcard)
	type casbinRule struct {
		Ptype string `gorm:"column:ptype"`
		V0    string `gorm:"column:v0"`
		V1    string `gorm:"column:v1"`
		V2    string `gorm:"column:v2"`
		V3    string `gorm:"column:v3"`
		V4    string `gorm:"column:v4"`
		V5    string `gorm:"column:v5"`
	}
	casbinRules := []casbinRule{
		{Ptype: "g", V0: fmt.Sprintf("user:%d", adminUser.ID), V1: fmt.Sprintf("role:%d", superRole.ID)},
		{Ptype: "p", V0: fmt.Sprintf("role:%d", superRole.ID), V1: "*", V2: "access"},
	}
	for _, rule := range casbinRules {
		if err := db.Table("hyadmin_casbin_rules").
			Clauses(clause.OnConflict{DoNothing: true}).
			Create(&rule).Error; err != nil {
			return fmt.Errorf("insert casbin rule ptype=%s v0=%s: %w", rule.Ptype, rule.V0, err)
		}
		fmt.Printf("  casbin: ptype=%s v0=%s v1=%s v2=%s\n",
			rule.Ptype, rule.V0, rule.V1, rule.V2)
	}

	// ── 7. Application settings ────────────────────────────────
	type setting struct {
		Key         string
		Value       string
		Type        string
		GroupName   string
		Description string
		IsPublic    bool
	}
	settings := []setting{
		{"auth.session.expire_hours", "24", "integer", "auth", "Session 有效時數", false},
		{"auth.password.min_length", "8", "integer", "auth", "密碼最短長度", false},
		{"auth.password.require_uppercase", "true", "boolean", "auth", "密碼須包含大寫字母", false},
		{"audit.log.retention_days", "90", "integer", "audit", "稽核日誌保留天數", false},
		{"ui.platform_name", "HySP Admin", "string", "ui", "平台顯示名稱", true},
		{"ui.logo_url", "", "string", "ui", "Logo URL", true},
	}
	for _, s := range settings {
		row := map[string]interface{}{
			"key":         s.Key,
			"value":       s.Value,
			"type":        s.Type,
			"group_name":  s.GroupName,
			"description": s.Description,
			"is_public":   s.IsPublic,
		}
		if err := db.Table("hyadmin_settings").
			Clauses(clause.OnConflict{DoNothing: true}).
			Create(row).Error; err != nil {
			return fmt.Errorf("upsert setting %q: %w", s.Key, err)
		}
		fmt.Printf("  setting: %s=%s\n", s.Key, s.Value)
	}

	return nil
}
