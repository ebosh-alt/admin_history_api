package repository

import (
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func Module(options ...fx.Option) fx.Option {
	base := []fx.Option{
		fx.Decorate(func(log *zap.Logger) *zap.Logger {
			return log.Named("repository")
		}),
	}
	base = append(base, options...)
	return fx.Module("repository", base...)
}
