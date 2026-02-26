//go:build ignore

// loader is a helper program invoked by the Atlas CLI to extract the desired
// database schema from GORM models.
//
//	atlas migrate diff --env local
package main

import (
	"fmt"
	"io"
	"os"

	"ariga.io/atlas-provider-gorm/gormschema"
	"github.com/hysp/hyadmin-api/internal/adminuser"
	"github.com/hysp/hyadmin-api/internal/auditlog"
	"github.com/hysp/hyadmin-api/internal/database"
	"github.com/hysp/hyadmin-api/internal/feature"
	"github.com/hysp/hyadmin-api/internal/pbmodule"
	"github.com/hysp/hyadmin-api/internal/permission"
	"github.com/hysp/hyadmin-api/internal/role"
	"github.com/hysp/hyadmin-api/internal/tenant"
)

func main() {
	// Register all GORM models that belong to the admin DB.
	stmts, err := gormschema.New("postgres").Load(
		&tenant.Tenant{},
		&database.TenantDBConfig{},
		&adminuser.AdminUser{},
		&role.Role{},
		&role.UserRole{},
		&pbmodule.PlatformModule{},
		&feature.Feature{},
		&permission.Permission{},
		&permission.RolePermission{},
		&auditlog.AuditLog{},
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load gorm schema: %v\n", err)
		os.Exit(1)
	}
	io.WriteString(os.Stdout, stmts)
}
