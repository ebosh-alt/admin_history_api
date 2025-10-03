package server

import (
	protos "admin_history/pkg/proto/gen/go"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"
	"io"
	"net/http"
)

var (
	unmarshalJSON = protojson.UnmarshalOptions{DiscardUnknown: true}
	marshalJSON   = protojson.MarshalOptions{EmitUnpopulated: true}
)

func (s *Server) GetQuestionnaire(c *gin.Context) {
	req := &protos.QuestionnaireRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		s.log.Error("failed to get questionnaire: ", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"message": "Не корректный запрос"})
		return
	}

	qProto, err := s.Usecase.GetQuestionnaire(c, req)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Не корректные данные"})
		return
	}
	c.JSON(http.StatusOK, qProto)
}
func (s *Server) QuestionnairesList(c *gin.Context) {
	req := &protos.QuestionnairesListRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		s.log.Error("failed to get questionnaires: ", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"message": "Не корректный запрос"})
		return
	}

	qProto, err := s.Usecase.GetQuestionnairesList(c, req)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Не корректные данные"})
		return
	}
	c.JSON(http.StatusOK, qProto)

}

func (s *Server) UpdateQuestionnaire(c *gin.Context) {
	body, _ := io.ReadAll(c.Request.Body)
	var req protos.UpdateQuestionnaireRequest
	if err := unmarshalJSON.Unmarshal(body, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid JSON"})
		return
	}
	resp, err := s.Usecase.UpdateQuestionnaire(c, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	b, _ := marshalJSON.Marshal(resp)
	c.Data(http.StatusOK, "application/json", b)
}

var _ InterfaceQuestionnaireServer = (*Server)(nil)
