package server

import (
	"github.com/gin-gonic/gin"
)

type InterfaceServer interface {
	InterfaceUserServer
	InterfaceQuestionnaireServer
	InterfacePhotoServer
	InterfaceVideoServer
	InterfaceChatServer
	InterfaceStatisticServer
	InterfacePromoCodeServer
	InterfaceReviewServer
}

type InterfaceUserServer interface {
	GetUser(c *gin.Context)
	UsersList(c *gin.Context)
	UpdateUser(c *gin.Context)
}

type InterfaceQuestionnaireServer interface {
	QuestionnairesList(c *gin.Context)
	GetQuestionnaire(c *gin.Context)
	UpdateQuestionnaire(c *gin.Context)
	SubmitQuestionnaireMedia(c *gin.Context)
}

type InterfacePhotoServer interface {
	GetPhotosQuestionnaire(c *gin.Context)
	UploadPhoto(c *gin.Context)
}

type InterfaceVideoServer interface {
	GetVideosQuestionnaire(c *gin.Context)
	UploadVideo(c *gin.Context)
}

type InterfaceChatServer interface {
	GetChat(c *gin.Context)
	ChatsList(c *gin.Context)
}

type InterfaceStatisticServer interface {
	GetStatistics(c *gin.Context)
}

type InterfacePromoCodeServer interface {
	GetPromoCode(c *gin.Context)
	PromoCodesList(c *gin.Context)
	CreatePromoCode(c *gin.Context)
	UpdatePromoCode(c *gin.Context)
}

type InterfaceReviewServer interface {
	GetReview(c *gin.Context)
	ReviewsList(c *gin.Context)
}
