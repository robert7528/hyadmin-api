package app

import (
	"github.com/hysp/hyadmin-api/internal/config"
	"github.com/hysp/hyadmin-api/internal/database"
	"github.com/hysp/hyadmin-api/internal/health"
	"github.com/hysp/hyadmin-api/internal/logger"
	"github.com/hysp/hyadmin-api/internal/module"
	"github.com/hysp/hyadmin-api/internal/server"
	"github.com/hysp/hyadmin-api/internal/tenant"
	"go.uber.org/fx"
)

func Run() error {
	app := fx.New(
		fx.Provide(
			config.Load,
			logger.New,
			database.Connect,
			server.New,
			module.NewRegistry,
			tenant.NewRepository,
			tenant.NewService,
			tenant.NewHandler,
			health.NewHandler,
		),
		fx.Invoke(server.RegisterRoutes),
		fx.Invoke(server.Start),
	)
	app.Run()
	return nil
}
