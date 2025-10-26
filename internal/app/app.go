package app

import (
	"admin_history/internal/repository/postgres"
	usecase "admin_history/internal/usecase"
	"context"

	"admin_history/config"
	"admin_history/internal/delivery/http/middleware"
	"admin_history/internal/delivery/server"
	"admin_history/internal/repository"
	"admin_history/internal/storage"
	"admin_history/pkg/telegram"

	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
)

func New() *fx.App {
	return fx.New(
		repository.Module(postgres.Module()),
		usecase.Module(),
		server.New(),
		// --- Provide all dependencies ---
		fx.Provide(
			// базовые
			context.Background,
			config.NewConfig,
			zap.NewDevelopment,

			// Usecase и его интерфейс
			storage.NewFS,
			telegram.NewClient,
			// HTTP-мiddleware и сервер
			middleware.NewMiddleware,
		),

		// --- Use Zap logger for Fx events ---
		fx.WithLogger(func(log *zap.Logger) fxevent.Logger {
			return &fxevent.ZapLogger{Logger: log}
		}),
	)
}
