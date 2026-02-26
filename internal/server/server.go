package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/casbin/casbin/v2"
	"github.com/gin-gonic/gin"
	"github.com/hysp/hyadmin-api/internal/adminuser"
	"github.com/hysp/hyadmin-api/internal/auditlog"
	"github.com/hysp/hyadmin-api/internal/auth"
	"github.com/hysp/hyadmin-api/internal/config"
	"github.com/hysp/hyadmin-api/internal/database"
	"github.com/hysp/hyadmin-api/internal/feature"
	"github.com/hysp/hyadmin-api/internal/health"
	"github.com/hysp/hyadmin-api/internal/middleware"
	"github.com/hysp/hyadmin-api/internal/pbmodule"
	"github.com/hysp/hyadmin-api/internal/permission"
	"github.com/hysp/hyadmin-api/internal/role"
	"github.com/hysp/hyadmin-api/internal/tenant"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Server struct {
	engine *gin.Engine
	cfg    *config.Config
	log    *zap.Logger
}

func New(cfg *config.Config, log *zap.Logger) *Server {
	gin.SetMode(cfg.Server.Mode)
	engine := gin.New()
	return &Server{engine: engine, cfg: cfg, log: log}
}

// RouteParams groups all handler dependencies for fx injection.
type RouteParams struct {
	fx.In
	Server     *Server
	DB         *gorm.DB
	Health     *health.Handler
	Tenant     *tenant.Handler
	Module     *pbmodule.Handler
	Feature    *feature.Handler
	AdminUser  *adminuser.Handler
	Role       *role.Handler
	RoleSvc    *role.Service
	Permission *permission.Handler
	Auth       *auth.Handler
	AuthSvc    *auth.Service
	AuditLog   *auditlog.Handler
	Enforcer   *casbin.Enforcer
	DBManager  *database.DBManager
}

func RegisterRoutes(p RouteParams) {
	r := p.Server.engine
	r.Use(middleware.Recovery(p.Server.log))

	api := r.Group("/api/v1")

	// ── Public routes (no JWT) ──────────────────────────────────────────
	api.GET("/health", p.Health.Check)
	api.POST("/auth/login", p.Auth.Login)
	api.POST("/auth/logout", p.Auth.Logout)

	// ── JWT-protected routes ────────────────────────────────────────────
	protected := api.Group("")
	protected.Use(middleware.AuthMiddleware(p.AuthSvc))
	protected.Use(middleware.PermissionLoaderMiddleware(p.RoleSvc))
	{
		// User-facing: modules & features (filtered by permissions)
		protected.GET("/modules", p.Module.ListForUser)
		protected.GET("/features", p.Feature.ListByModule)

		// Current user's permission codes
		protected.GET("/permissions/me", func(c *gin.Context) {
			codes := middleware.GetPermissionCodes(c)
			c.JSON(http.StatusOK, gin.H{"permissions": codes})
		})

		// Profile routes (use JWT claims for user identity)
		profile := protected.Group("/profile")
		{
			profile.PUT("/display-name", func(c *gin.Context) {
				claims := middleware.GetClaims(c)
				if claims == nil {
					c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
					return
				}
				var req struct {
					DisplayName string `json:"display_name" binding:"required"`
				}
				if err := c.ShouldBindJSON(&req); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}
				if err := p.AdminUser.UpdateSelf(c, claims.UserID, req.DisplayName); err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
				c.JSON(http.StatusOK, gin.H{"message": "updated"})
			})
			profile.PUT("/password", func(c *gin.Context) {
				claims := middleware.GetClaims(c)
				if claims == nil {
					c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
					return
				}
				var req adminuser.ChangePasswordRequest
				if err := c.ShouldBindJSON(&req); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}
				if err := p.AdminUser.ChangeSelfPassword(c, claims.UserID, &req); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}
				c.JSON(http.StatusOK, gin.H{"message": "password updated"})
			})
		}

		// Tenant CRUD (admin)
		tenants := protected.Group("/tenants")
		tenants.Use(middleware.TenantMiddleware())
		{
			tenants.GET("", p.Tenant.List)
			tenants.POST("", p.Tenant.Create)
			tenants.GET("/:id", p.Tenant.Get)
			tenants.PUT("/:id", p.Tenant.Update)
			tenants.DELETE("/:id", p.Tenant.Delete)
		}

		// Admin routes
		admin := protected.Group("/admin")
		admin.Use(middleware.PermissionMiddleware(p.Enforcer))
		admin.Use(auditlog.AuditMiddleware(p.DB))
		{
			// Modules
			mods := admin.Group("/modules")
			{
				mods.GET("", p.Module.List)
				mods.POST("", p.Module.Create)
				mods.GET("/:id", p.Module.Get)
				mods.PUT("/:id", p.Module.Update)
				mods.DELETE("/:id", p.Module.Delete)
				// Features under a module
				mods.GET("/:moduleId/features", p.Feature.ListByModule)
				mods.POST("/:moduleId/features", p.Feature.Create)
			}

			// Features (individual)
			feats := admin.Group("/features")
			{
				feats.GET("/:id", p.Feature.Get)
				feats.PUT("/:id", p.Feature.Update)
				feats.DELETE("/:id", p.Feature.Delete)
				// Permissions under a feature
				feats.GET("/:id/permissions", p.Permission.ListByFeature)
				feats.POST("/:id/permissions", p.Permission.Create)
				feats.POST("/:id/permissions/batch", p.Permission.BatchCreate)
			}

			// Permissions (individual)
			perms := admin.Group("/permissions")
			{
				perms.PUT("/:id", p.Permission.Update)
				perms.DELETE("/:id", p.Permission.Delete)
			}

			// Users
			users := admin.Group("/users")
			{
				users.GET("", p.AdminUser.List)
				users.POST("", p.AdminUser.Create)
				users.GET("/:id", p.AdminUser.Get)
				users.PUT("/:id", p.AdminUser.Update)
				users.PUT("/:id/password", p.AdminUser.ChangePassword)
				users.DELETE("/:id", p.AdminUser.Delete)
			}

			// Audit logs
			admin.GET("/audit-logs", p.AuditLog.List)

			// Roles
			roles := admin.Group("/roles")
			{
				roles.GET("", p.Role.List)
				roles.POST("", p.Role.Create)
				roles.GET("/:id", p.Role.Get)
				roles.PUT("/:id", p.Role.Update)
				roles.DELETE("/:id", p.Role.Delete)
				roles.GET("/:id/permissions", p.Role.GetPermissions)
				roles.PUT("/:id/permissions", p.Role.AssignPermissions)
				roles.PUT("/:id/users", p.Role.AssignUsers)
			}
		}

		// Business data routes with tenant DB
		data := protected.Group("/data")
		data.Use(middleware.TenantMiddleware())
		data.Use(middleware.TenantDBMiddleware(p.DBManager))
		{
			_ = data
		}
	}
}

func Start(lc fx.Lifecycle, s *Server) {
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", s.cfg.Server.Port),
		Handler: s.engine,
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			s.log.Info("starting server", zap.String("addr", srv.Addr))
			go func() {
				if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					s.log.Fatal("server error", zap.Error(err))
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			s.log.Info("stopping server")
			return srv.Shutdown(ctx)
		},
	})
}
