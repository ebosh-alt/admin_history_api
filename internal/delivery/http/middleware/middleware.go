package middleware

import (
	"admin_history/config"
	"go.uber.org/zap"
)

type Middleware struct {
	cfg   *config.Config
	log   *zap.Logger
	roles map[string]int
}

func NewMiddleware(cfg *config.Config, log *zap.Logger) (*Middleware, error) {
	return &Middleware{
		cfg: cfg,
		log: log,
		// repo: repository,
	}, nil
}
