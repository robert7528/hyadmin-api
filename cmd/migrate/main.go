package main

import (
	"context"
	"fmt"
	"log"

	"github.com/hysp/hyadmin-api/internal/config"
	"github.com/hysp/hyadmin-api/internal/database"
	"github.com/spf13/cobra"
)

const (
	adminMigrationsDir  = "migrations/admin"
	tenantMigrationsDir = "migrations/tenant"
)

func main() {
	root := &cobra.Command{
		Use:   "migrate",
		Short: "Run Atlas database migrations",
	}
	root.AddCommand(adminCmd(), tenantCmd(), allTenantsCmd())
	if err := root.Execute(); err != nil {
		log.Fatal(err)
	}
}

// migrate admin — apply migrations/admin against the admin DB
func adminCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "admin",
		Short: "Apply admin DB migrations",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := config.Load()
			db, err := database.Connect(cfg)
			if err != nil {
				return fmt.Errorf("connect admin DB: %w", err)
			}
			fmt.Println("Applying admin migrations...")
			if err := database.MigrateAdmin(context.Background(), db, adminMigrationsDir); err != nil {
				return err
			}
			fmt.Println("Admin migrations applied successfully.")
			return nil
		},
	}
}

// migrate tenant --code TENANT_CODE — apply migrations/tenant for one tenant
func tenantCmd() *cobra.Command {
	var code string
	cmd := &cobra.Command{
		Use:   "tenant",
		Short: "Apply tenant DB migrations for a specific tenant",
		RunE: func(cmd *cobra.Command, args []string) error {
			if code == "" {
				return fmt.Errorf("--code is required")
			}
			return applyTenantMigration(code)
		},
	}
	cmd.Flags().StringVar(&code, "code", "", "Tenant code (required)")
	return cmd
}

// migrate all-tenants — apply migrations/tenant for every tenant in admin DB
func allTenantsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "all-tenants",
		Short: "Apply tenant DB migrations for all registered tenants",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := config.Load()
			adminDB, err := database.Connect(cfg)
			if err != nil {
				return fmt.Errorf("connect admin DB: %w", err)
			}

			var configs []database.TenantDBConfig
			if err := adminDB.Find(&configs).Error; err != nil {
				return fmt.Errorf("list tenant DB configs: %w", err)
			}

			mgr := database.NewManager(adminDB)
			ctx := context.Background()
			var failed []string

			for _, cfg := range configs {
				fmt.Printf("  → tenant %s ...", cfg.TenantCode)
				tenantDB, err := mgr.GetDB(cfg.TenantCode)
				if err != nil {
					fmt.Printf(" ERROR: %v\n", err)
					failed = append(failed, cfg.TenantCode)
					continue
				}
				if err := database.MigrateTenant(ctx, tenantDB, tenantMigrationsDir, cfg.Schema); err != nil {
					fmt.Printf(" ERROR: %v\n", err)
					failed = append(failed, cfg.TenantCode)
					continue
				}
				fmt.Println(" ok")
			}

			if len(failed) > 0 {
				return fmt.Errorf("migrations failed for tenants: %v", failed)
			}
			fmt.Printf("All %d tenant migrations applied successfully.\n", len(configs))
			return nil
		},
	}
}

func applyTenantMigration(tenantCode string) error {
	cfg := config.Load()
	adminDB, err := database.Connect(cfg)
	if err != nil {
		return fmt.Errorf("connect admin DB: %w", err)
	}

	var tenantCfg database.TenantDBConfig
	if err := adminDB.Where("tenant_code = ?", tenantCode).First(&tenantCfg).Error; err != nil {
		return fmt.Errorf("tenant %q not found: %w", tenantCode, err)
	}

	mgr := database.NewManager(adminDB)
	tenantDB, err := mgr.GetDB(tenantCode)
	if err != nil {
		return fmt.Errorf("connect tenant DB: %w", err)
	}

	fmt.Printf("Applying tenant migrations for %q ...\n", tenantCode)
	if err := database.MigrateTenant(context.Background(), tenantDB, tenantMigrationsDir, tenantCfg.Schema); err != nil {
		return err
	}
	fmt.Printf("Tenant %q migrations applied successfully.\n", tenantCode)
	return nil
}
