package server

import (
	"admin_history/config"
	"admin_history/internal/delivery/http/middleware"
	"admin_history/internal/usecase"
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func NewServer(logger *zap.Logger, cfg *config.Config, uc usecase.InterfaceUsecase, middleware *middleware.Middleware) (*Server, error) {
	return &Server{
		log:        logger,
		cfg:        cfg,
		serv:       gin.Default(),
		Usecase:    uc,
		middleware: middleware,
	}, nil
}

func New() fx.Option {
	return fx.Module("NewServer",
		fx.Provide(
			NewServer,
		),
		fx.Invoke(
			func(lc fx.Lifecycle, s *Server) {
				lc.Append(fx.Hook{
					OnStart: s.OnStart,
					OnStop:  s.OnStop,
				})
			},
		),
		fx.Decorate(func(log *zap.Logger) *zap.Logger {
			return log.Named("server")
		}),
	)
}
