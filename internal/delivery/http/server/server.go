package server

import (
	"admin_history/config"
	"admin_history/internal/delivery/http/middleware"
	"admin_history/internal/usecase"
	"context"
	_ "fmt"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Server struct {
	log        *zap.Logger
	cfg        *config.Config
	serv       *gin.Engine
	Usecase    usecase.InterfaceUsecase
	middleware *middleware.Middleware
}

func (s *Server) OnStart(_ context.Context) error {
	s.CreateController()
	go func() {
		s.log.Debug("server started")
		if err := s.serv.Run(s.cfg.Server.Host + ":" + s.cfg.Server.Port); err != nil {
			s.log.Error("failed to server: " + err.Error())
		}
		return
	}()
	return nil
}

func (s *Server) OnStop(_ context.Context) error {
	s.log.Debug("stop server")
	return nil
}
