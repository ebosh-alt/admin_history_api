package server

import (
	_ "admin_history/internal/delivery/http/server/docs"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func (s *Server) CreateController() {
	s.serv.Use(s.middleware.CORSMiddleware)
	s.serv.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	api := s.serv.Group("/api")

	{
		apiUser := api.Group("/users")
		{
			apiUser.GET("/:id", s.GetUser)
			apiUser.GET("/", s.UsersList)
			apiUser.POST("/update", s.UpdateUser)
		}
	}

	{
		apiQuestionnaire := api.Group("/questionnaires")
		{
			apiQuestionnaire.GET("/:id", s.GetQuestionnaire)
			apiQuestionnaire.GET("", s.QuestionnairesList)
			apiQuestionnaire.POST("/update", s.UpdateQuestionnaire)
		}
	}

	{
		apiStatistics := api.Group("/statistics")
		{
			apiStatistics.GET("/chat", s.GetStatistics)
		}
	}

}
