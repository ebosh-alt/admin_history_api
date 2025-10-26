package server

import (
	_ "admin_history/internal/delivery/server/docs"
	"admin_history/internal/delivery/server/domain/photo"
	"admin_history/internal/delivery/server/domain/promo"
	"admin_history/internal/delivery/server/domain/questionnaire"
	"admin_history/internal/delivery/server/domain/review"
	"admin_history/internal/delivery/server/domain/user"
	"admin_history/internal/delivery/server/domain/video"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func (s *Server) CreateController() {
	s.serv.Use(s.middleware.CORSMiddleware)
	s.serv.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	api := s.serv.Group("/api")

	userHandler := user.New(s.log.Named("user"), s.usecase)
	userHandler.RegisterRoutes(api.Group("/users"))

	questionnaireHandler := questionnaire.New(s.log.Named("questionnaire"), s.usecase)
	questionnaireHandler.RegisterRoutes(api.Group("/questionnaires"))

	photoHandler := photo.New(s.log.Named("photo"), s.usecase)
	photoHandler.RegisterRoutes(api.Group("/photos"))

	videoHandler := video.New(s.log.Named("video"), s.usecase)
	videoHandler.RegisterRoutes(api.Group("/videos"))

	promoHandler := promo.New(s.log.Named("promo"), s.usecase)
	promoHandler.RegisterRoutes(api.Group("/promo-codes"))

	reviewHandler := review.New(s.log.Named("review"), s.usecase)
	reviewHandler.RegisterRoutes(api.Group("/reviews"))
}
