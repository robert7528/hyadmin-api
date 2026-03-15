package app

import (
	"github.com/casbin/casbin/v2"
	"github.com/hysp/hyadmin-api/internal/adminuser"
	"github.com/hysp/hyadmin-api/internal/auditlog"
	localauth "github.com/hysp/hyadmin-api/internal/auth"
	"github.com/hysp/hyadmin-api/internal/feature"
	"github.com/hysp/hyadmin-api/internal/health"
	"github.com/hysp/hyadmin-api/internal/pbmodule"
	"github.com/hysp/hyadmin-api/internal/permission"
	"github.com/hysp/hyadmin-api/internal/role"
	"github.com/hysp/hyadmin-api/internal/server"
	"github.com/hysp/hyadmin-api/internal/tenant"
	coreauth "github.com/robert7528/hycore/auth"
	"github.com/robert7528/hycore/casbinx"
	"github.com/robert7528/hycore/config"
	"github.com/robert7528/hycore/crypto"
	"github.com/robert7528/hycore/database"
	"github.com/robert7528/hycore/logger"
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
				return casbinx.NewEnforcer(db, "configs/rbac_model.conf", "hyadmin_casbin_rules")
			},

			// AdminUser domain
			adminuser.NewRepository,
			adminuser.NewService,
			adminuser.NewHandler,

			// Auth domain
			func(cfg *config.Config, userSvc *adminuser.Service) *localauth.LocalProvider {
				return localauth.NewLocalProvider(userSvc, cfg.JWT.ExpiryHours)
			},
			func(cfg *config.Config, lp *localauth.LocalProvider) *coreauth.Service {
				return coreauth.NewService(cfg, lp)
			},
			coreauth.NewHandler,

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
