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
	"github.com/robert7528/hycore/config"
	"github.com/robert7528/hycore/crypto"
	"github.com/robert7528/hycore/database"
	"github.com/hysp/hyadmin-api/internal/feature"
	"github.com/hysp/hyadmin-api/internal/pbmodule"
	"github.com/hysp/hyadmin-api/internal/permission"
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

type casbinRule struct {
	Ptype string `gorm:"column:ptype"`
	V0    string `gorm:"column:v0"`
	V1    string `gorm:"column:v1"`
	V2    string `gorm:"column:v2"`
	V3    string `gorm:"column:v3"`
	V4    string `gorm:"column:v4"`
	V5    string `gorm:"column:v5"`
}

func run(db *gorm.DB, enc crypto.Encryptor) error {
	now := time.Now()

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
		{
			Name:        "cert",
			DisplayName: "憑證管理",
			I18n:        `{"zh-TW":"憑證管理","en":"Certificates"}`,
			Route:       "cert",
			ApiURL:      "/hycert-api",
			SortOrder:   10,
			Enabled:     true,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
	}
	moduleMap := make(map[string]uint) // name → id
	for i := range modules {
		m := &modules[i]
		if err := db.Where(pbmodule.PlatformModule{Name: m.Name}).
			FirstOrCreate(m).Error; err != nil {
			return fmt.Errorf("upsert module %q: %w", m.Name, err)
		}
		moduleMap[m.Name] = m.ID
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

	// ── 6. Casbin base policies ────────────────────────────────
	baseCasbinRules := []casbinRule{
		{Ptype: "g", V0: fmt.Sprintf("user:%d", adminUser.ID), V1: fmt.Sprintf("role:%d", superRole.ID)},
		{Ptype: "p", V0: fmt.Sprintf("role:%d", superRole.ID), V1: "*", V2: "access"},
	}
	for _, r := range baseCasbinRules {
		if err := db.Table("hyadmin_casbin_rules").
			Clauses(clause.OnConflict{DoNothing: true}).
			Create(&r).Error; err != nil {
			return fmt.Errorf("insert casbin rule ptype=%s v0=%s: %w", r.Ptype, r.V0, err)
		}
		fmt.Printf("  casbin: ptype=%s v0=%s v1=%s v2=%s\n", r.Ptype, r.V0, r.V1, r.V2)
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

	// ── 8. Features ────────────────────────────────────────────
	type featureSeed struct {
		ModuleCode  string
		Name        string
		DisplayName string
		I18n        string
		Path        string
		SortOrder   int
	}
	featureSeeds := []featureSeed{
		{"tenants", "tenant-list", "租戶列表", `{"zh-TW":"租戶列表","en":"Tenant List"}`, "", 1},
		{"users", "user-list", "使用者列表", `{"zh-TW":"使用者列表","en":"User List"}`, "", 1},
		{"rbac", "role-list", "角色管理", `{"zh-TW":"角色管理","en":"Role Management"}`, "/roles", 1},
		{"audit", "audit-log", "稽核日誌", `{"zh-TW":"稽核日誌","en":"Audit Log"}`, "", 1},
		{"settings", "settings", "系統設定", `{"zh-TW":"系統設定","en":"System Settings"}`, "", 1},
		// cert
		{"cert", "cert-toolbox", "工具箱", `{"zh-TW":"工具箱","en":"Toolbox"}`, "/toolbox", 1},
		{"cert", "cert-list", "憑證列表", `{"zh-TW":"憑證列表","en":"Certificates"}`, "/list", 2},
		{"cert", "cert-deployments", "部署目標", `{"zh-TW":"部署目標","en":"Deployments"}`, "/deployments", 3},
	}
	featureMap := make(map[string]feature.Feature) // name → feature
	for _, fs := range featureSeeds {
		f := feature.Feature{
			ModuleID:    moduleMap[fs.ModuleCode],
			Name:        fs.Name,
			DisplayName: fs.DisplayName,
			I18n:        fs.I18n,
			Path:        fs.Path,
			SortOrder:   fs.SortOrder,
			Enabled:     true,
			CreatedAt:   now,
			UpdatedAt:   now,
		}
		if err := db.Where(feature.Feature{Name: fs.Name}).
			FirstOrCreate(&f).Error; err != nil {
			return fmt.Errorf("upsert feature %q: %w", fs.Name, err)
		}
		featureMap[f.Name] = f
		fmt.Printf("  feature: id=%d name=%s\n", f.ID, f.Name)
	}

	// ── 9. Permissions ─────────────────────────────────────────
	type permSeed struct {
		FeatureName string
		Code        string
		Name        string
		I18n        string
		Type        string // menu|button|api
		SortOrder   int
	}
	permSeeds := []permSeed{
		// tenants
		{"tenant-list", "tenants.list.view", "租戶列表頁面", `{"zh-TW":"租戶列表頁面","en":"Tenant List"}`, "menu", 1},
		{"tenant-list", "tenants.list.create", "新增租戶", `{"zh-TW":"新增租戶","en":"Create Tenant"}`, "button", 2},
		{"tenant-list", "tenants.list.update", "編輯租戶", `{"zh-TW":"編輯租戶","en":"Edit Tenant"}`, "button", 3},
		{"tenant-list", "tenants.list.delete", "刪除租戶", `{"zh-TW":"刪除租戶","en":"Delete Tenant"}`, "button", 4},
		// users
		{"user-list", "users.list.view", "使用者列表頁面", `{"zh-TW":"使用者列表頁面","en":"User List"}`, "menu", 1},
		{"user-list", "users.list.create", "新增使用者", `{"zh-TW":"新增使用者","en":"Create User"}`, "button", 2},
		{"user-list", "users.list.update", "編輯使用者", `{"zh-TW":"編輯使用者","en":"Edit User"}`, "button", 3},
		{"user-list", "users.list.delete", "刪除使用者", `{"zh-TW":"刪除使用者","en":"Delete User"}`, "button", 4},
		{"user-list", "users.list.change_password", "修改密碼", `{"zh-TW":"修改密碼","en":"Change Password"}`, "button", 5},
		// rbac
		{"role-list", "rbac.roles.view", "角色管理頁面", `{"zh-TW":"角色管理頁面","en":"Role Management"}`, "menu", 1},
		{"role-list", "rbac.roles.create", "新增角色", `{"zh-TW":"新增角色","en":"Create Role"}`, "button", 2},
		{"role-list", "rbac.roles.update", "編輯角色", `{"zh-TW":"編輯角色","en":"Edit Role"}`, "button", 3},
		{"role-list", "rbac.roles.delete", "刪除角色", `{"zh-TW":"刪除角色","en":"Delete Role"}`, "button", 4},
		{"role-list", "rbac.roles.assign", "指派權限", `{"zh-TW":"指派權限","en":"Assign Permissions"}`, "button", 5},
		// audit
		{"audit-log", "audit.logs.view", "稽核日誌頁面", `{"zh-TW":"稽核日誌頁面","en":"Audit Log"}`, "menu", 1},
		{"audit-log", "audit.logs.export", "匯出稽核日誌", `{"zh-TW":"匯出稽核日誌","en":"Export Logs"}`, "button", 2},
		// settings
		{"settings", "settings.view", "系統設定頁面", `{"zh-TW":"系統設定頁面","en":"System Settings"}`, "menu", 1},
		{"settings", "settings.update", "修改系統設定", `{"zh-TW":"修改系統設定","en":"Update Settings"}`, "button", 2},
		// cert
		{"cert-toolbox", "cert.toolbox.view", "工具箱頁面", `{"zh-TW":"工具箱頁面","en":"Toolbox"}`, "menu", 1},
		{"cert-list", "cert.list.view", "憑證列表頁面", `{"zh-TW":"憑證列表頁面","en":"Certificate List"}`, "menu", 1},
		{"cert-list", "cert.list.create", "新增憑證", `{"zh-TW":"新增憑證","en":"Create Certificate"}`, "button", 2},
		{"cert-list", "cert.list.delete", "刪除憑證", `{"zh-TW":"刪除憑證","en":"Delete Certificate"}`, "button", 3},
	}

	seededPerms := make([]permission.Permission, 0, len(permSeeds))
	for _, ps := range permSeeds {
		f := featureMap[ps.FeatureName]
		p := permission.Permission{
			FeatureID: f.ID,
			Code:      ps.Code,
			Name:      ps.Name,
			I18n:      ps.I18n,
			Type:      ps.Type,
			SortOrder: ps.SortOrder,
			CreatedAt: now,
			UpdatedAt: now,
		}
		if err := db.Where(permission.Permission{Code: ps.Code}).
			FirstOrCreate(&p).Error; err != nil {
			return fmt.Errorf("upsert permission %q: %w", ps.Code, err)
		}
		seededPerms = append(seededPerms, p)
		fmt.Printf("  permission: id=%d code=%s\n", p.ID, p.Code)
	}

	// ── 10. Assign permissions to super_admin ──────────────────
	for _, perm := range seededPerms {
		rp := permission.RolePermission{
			RoleID:       superRole.ID,
			PermissionID: perm.ID,
		}
		if err := db.Clauses(clause.OnConflict{DoNothing: true}).
			Create(&rp).Error; err != nil {
			return fmt.Errorf("upsert role_permission for %s: %w", perm.Code, err)
		}
		// Add specific casbin p policy so GetPermissionCodesForUser returns real codes
		rule := casbinRule{
			Ptype: "p",
			V0:    fmt.Sprintf("role:%d", superRole.ID),
			V1:    perm.Code,
			V2:    "access",
		}
		if err := db.Table("hyadmin_casbin_rules").
			Clauses(clause.OnConflict{DoNothing: true}).
			Create(&rule).Error; err != nil {
			return fmt.Errorf("insert casbin p rule for %s: %w", perm.Code, err)
		}
	}
	fmt.Printf("  role_permissions + casbin: assigned %d permissions to role %s\n",
		len(seededPerms), superRole.Name)

	return nil
}
