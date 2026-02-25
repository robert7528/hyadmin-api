package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hysp/hyadmin-api/internal/config"
	"github.com/hysp/hyadmin-api/internal/health"
	"github.com/hysp/hyadmin-api/internal/middleware"
	"github.com/hysp/hyadmin-api/internal/module"
	"github.com/hysp/hyadmin-api/internal/tenant"
	"go.uber.org/fx"
	"go.uber.org/zap"
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
	Server   *Server
	Health   *health.Handler
	Tenant   *tenant.Handler
	Registry *module.Registry
}

func RegisterRoutes(p RouteParams) {
	r := p.Server.engine
	r.Use(middleware.Recovery(p.Server.log))
	r.Use(middleware.TenantMiddleware())
	r.Use(middleware.AuthMiddleware())

	api := r.Group("/api/v1")
	{
		api.GET("/health", p.Health.Check)
		api.GET("/modules", p.Registry.ListModules)

		tenants := api.Group("/tenants")
		{
			tenants.GET("", p.Tenant.List)
			tenants.POST("", p.Tenant.Create)
			tenants.GET("/:id", p.Tenant.Get)
			tenants.PUT("/:id", p.Tenant.Update)
			tenants.DELETE("/:id", p.Tenant.Delete)
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
