package repository

import (
	"admin_history/internal/repository/postgres"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func New() fx.Option {
	return fx.Module("repository",
		fx.Provide(
			postgres.NewRepository,
		),
		fx.Invoke(
			func(lc fx.Lifecycle, pg *postgres.Repository) {
				lc.Append(fx.Hook{
					OnStart: pg.OnStart,
					OnStop:  pg.OnStop,
				})
			},
		),
		fx.Decorate(func(log *zap.Logger) *zap.Logger {
			return log.Named("repository")
		}),
	)
}
