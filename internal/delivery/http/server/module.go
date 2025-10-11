package server

import (
	"admin_history/config"
	"admin_history/internal/delivery/http/middleware"
	"admin_history/internal/storage"
	"admin_history/internal/usecase"
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func NewServer(
	logger *zap.Logger,
	cfg *config.Config,
	uc usecase.InterfaceUsecase,
	middleware *middleware.Middleware,
	st *storage.FS,
) (*Server, error) {
	engine := gin.Default()
	if st != nil {
		if route := st.PublicRoute(); route != "" {
			if dir := st.PublicDir(); dir != "" {
				engine.Static(route, dir)
			}
		}
	}

	return &Server{
		log:        logger,
		cfg:        cfg,
		serv:       engine,
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
