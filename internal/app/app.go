// internal/app/app.go

package app

import (
	"admin_history/internal/delivery/http/handlers"
	"admin_history/internal/delivery/http/middleware"
	"admin_history/internal/delivery/http/server"
	"context"

	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"

	"admin_history/config"
	"admin_history/internal/repository/postgres"
	"admin_history/internal/usecase"
)

func New() *fx.App {
	return fx.New(
		// --- Provide all dependencies ---
		fx.Provide(
			// базовые
			context.Background,
			config.NewConfig,
			zap.NewDevelopment,

			// Postgres-репозиторий и его интерфейс
			postgres.NewRepository,
			func(r *postgres.Repository) postgres.InterfaceRepo {
				return r
			},

			// Usecase и его интерфейс
			usecase.NewUsecase,
			func(u *usecase.Usecase) usecase.InterfaceUsecase {
				return u
			},
			// HTTP-мiddleware и сервер
			middleware.NewMiddleware,
			handlers.NewServer,
		),
		// Lifecycle: сначала поднимаем репозиторий
		fx.Invoke(func(lc fx.Lifecycle, repo *postgres.Repository) {
			lc.Append(fx.Hook{
				OnStart: repo.OnStart,
				OnStop:  repo.OnStop,
			})
		}),
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
