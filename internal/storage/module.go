package storage

import (
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func New() fx.Option {
	return fx.Module(
		"storage",
		fx.Provide(
			NewFS,
		),
		fx.Invoke(
			func(lc fx.Lifecycle, fs *FS) {
				lc.Append(fx.Hook{
					OnStart: fs.OnStart,
					OnStop:  fs.OnStop,
				})
			},
		),
		fx.Decorate(func(log *zap.Logger) *zap.Logger {
			return log.Named("storage")
		}),
	)
}
