package app

import (
	"admin_history/config"
	"admin_history/internal/delivery/http/middleware"
	"admin_history/internal/delivery/http/server"
	"admin_history/internal/repository"
	"admin_history/internal/repository/postgres"
	"admin_history/internal/storage"
	"admin_history/internal/usecase"
	"admin_history/pkg/telegram"
	"context"

	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
)

func New() *fx.App {
	return fx.New(
		repository.Module(postgres.Module()),
		// --- Provide all dependencies ---
		fx.Provide(
			// базовые
			context.Background,
			config.NewConfig,
			zap.NewDevelopment,

			// Usecase и его интерфейс
			usecase.NewUsecase,
			func(u *usecase.Usecase) usecase.InterfaceUsecase {
				return u

			},
			storage.NewFS,
			telegram.NewClient,
			// HTTP-мiddleware и сервер
			middleware.NewMiddleware,
			server.NewServer,
		),
		// --- Hook server lifecycle ---
		fx.Invoke(func(lc fx.Lifecycle, srv *server.Server) {
			lc.Append(fx.Hook{
				OnStart: srv.OnStart,
				OnStop:  srv.OnStop,
			})
		}),

		// --- Use Zap logger for Fx events ---
		fx.WithLogger(func(log *zap.Logger) fxevent.Logger {
			return &fxevent.ZapLogger{Logger: log}
		}),
	)
}
