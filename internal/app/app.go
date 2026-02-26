package app

import (
	"github.com/casbin/casbin/v2"
	"github.com/hysp/hyadmin-api/internal/adminuser"
	"github.com/hysp/hyadmin-api/internal/auditlog"
	"github.com/hysp/hyadmin-api/internal/auth"
	"github.com/hysp/hyadmin-api/internal/casbinx"
	"github.com/hysp/hyadmin-api/internal/config"
	"github.com/hysp/hyadmin-api/internal/crypto"
	"github.com/hysp/hyadmin-api/internal/database"
	"github.com/hysp/hyadmin-api/internal/feature"
	"github.com/hysp/hyadmin-api/internal/health"
	"github.com/hysp/hyadmin-api/internal/logger"
	"github.com/hysp/hyadmin-api/internal/pbmodule"
	"github.com/hysp/hyadmin-api/internal/permission"
	"github.com/hysp/hyadmin-api/internal/role"
	"github.com/hysp/hyadmin-api/internal/server"
	"github.com/hysp/hyadmin-api/internal/tenant"
	"go.uber.org/fx"
	"gorm.io/gorm"
)

func Run() error {
	app := fx.New(
		fx.Provide(
			// Infrastructure
			config.Load,
			logger.New,
			database.Connect,
			database.NewManager,
			server.New,

			// Crypto (Tink)
			func(cfg *config.Config) (crypto.Encryptor, error) {
				return crypto.New(cfg.Tink.Keyset)
			},

			// Casbin enforcer
			func(db *gorm.DB) (*casbin.Enforcer, error) {
				return casbinx.NewEnforcer(db, "configs/rbac_model.conf")
			},

			// AdminUser domain
			adminuser.NewRepository,
			adminuser.NewService,
			adminuser.NewHandler,

			// Auth domain
			func(cfg *config.Config, userSvc *adminuser.Service) *auth.LocalProvider {
				return auth.NewLocalProvider(userSvc, cfg.JWT.ExpiryHours)
			},
			func(cfg *config.Config, lp *auth.LocalProvider) *auth.Service {
				return auth.NewService(cfg, lp)
			},
			auth.NewHandler,

			// Feature domain
			feature.NewRepository,
			feature.NewService,
			feature.NewHandler,

			// Permission domain
			permission.NewRepository,
			permission.NewService,
			permission.NewHandler,

			// PlatformModule domain
			pbmodule.NewRepository,
			pbmodule.NewService,
			pbmodule.NewHandler,

			// Role domain
			role.NewRepository,
			role.NewService,
			role.NewHandler,

			// Tenant domain
			tenant.NewRepository,
			tenant.NewService,
			tenant.NewHandler,

			// AuditLog
			auditlog.NewHandler,

			// Health
			health.NewHandler,
		),
		fx.Invoke(server.RegisterRoutes),
		fx.Invoke(server.Start),
	)
	app.Run()
	return nil
}
