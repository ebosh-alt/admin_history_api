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
			apiQuestionnaire.POST("/media", s.SubmitQuestionnaireMedia)
		}
	}

	{
		apiPhoto := api.Group("/photos")
		{
			apiPhoto.POST("/upload", s.UploadPhoto)
			apiPhoto.GET("", s.GetPhotosQuestionnaire)
		}

		apiVideo := api.Group("/videos")
		{
			apiVideo.POST("/upload", s.UploadVideo)
			apiVideo.GET("", s.GetVideosQuestionnaire)
		}
	}

	{
		apiPromoCode := api.Group("/promo-codes")
		{
			apiPromoCode.GET("/:id", s.GetPromoCode)
			apiPromoCode.GET("", s.PromoCodesList)
			apiPromoCode.POST("", s.CreatePromoCode)
			apiPromoCode.POST("/update", s.UpdatePromoCode)
		}
	}

	{
		apiReview := api.Group("/reviews")
		{
			apiReview.GET("/:id", s.GetReview)
			apiReview.GET("", s.ReviewsList)
		}
	}

}
