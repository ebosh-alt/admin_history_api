package connection

import (
	"admin_history/config"
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Params struct {
	fx.In

	Lifecycle fx.Lifecycle
	Cfg       *config.Config
	Logger    *zap.Logger
	Context   context.Context
}

func NewPool(p Params) (*pgxpool.Pool, error) {
	connectionURL := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		p.Cfg.Postgres.Host,
		p.Cfg.Postgres.Port,
		p.Cfg.Postgres.User,
		p.Cfg.Postgres.Password,
		p.Cfg.Postgres.DBName,
		p.Cfg.Postgres.SSLMode,
	)

	log := p.Logger.Named("postgres")
	log.Info("init pgx pool", zap.String("dsn", connectionURL))

	pool, err := pgxpool.New(p.Context, connectionURL)
	if err != nil {
		return nil, err
	}

	p.Lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			if err := pool.Ping(ctx); err != nil {
				return err
			}
			log.Info("postgres pool ready")
			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.Info("closing postgres pool")
			pool.Close()
			return nil
		},
	})

	return pool, nil
}
