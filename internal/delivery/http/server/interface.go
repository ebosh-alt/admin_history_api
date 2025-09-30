package server

import "github.com/gin-gonic/gin"

type InterfaceServer interface {
	InterfaceUserServer
	InterfaceQuestionnaireServer
	InterfaceChatServer
	InterfaceStatisticServer
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
}

type InterfaceChatServer interface {
	GetChat(c *gin.Context)
	ChatsList(c *gin.Context)
}

type InterfaceStatisticServer interface {
	GetStatistics(c *gin.Context)
}
