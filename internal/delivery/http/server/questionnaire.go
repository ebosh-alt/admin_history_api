package server

import (
	protos "admin_history/pkg/proto/gen/go"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
	"io"
	"net/http"
	"strconv"
	"time"
)

var (
	unmarshalJSON = protojson.UnmarshalOptions{DiscardUnknown: true}
	marshalJSON   = protojson.MarshalOptions{EmitUnpopulated: true}
)

func (s *Server) GetQuestionnaire(c *gin.Context) {
	req := &protos.QuestionnaireRequest{}
	//if err := c.ShouldBindJSON(&req); err != nil {
	//	s.log.Error("failed to get questionnaire: ", zap.Error(err))
	//	c.JSON(http.StatusBadRequest, gin.H{"message": "Не корректный запрос"})
	//	return
	//}
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": fmt.Sprintf("failed to unmarshar request: %v", err)})
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

	if v := c.Query("page"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 32); err == nil {
			req.Page = int32(n)
		}
	}
	if v := c.Query("limit"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 32); err == nil {
			req.Limit = int32(n)
		}
	}

	if v := c.Query("payment"); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			req.Payment = wrapperspb.Bool(b)
		}
	}
	if v := c.Query("status"); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			req.Status = wrapperspb.Bool(b)
		}
	}
	if v := c.Query("date_from"); v != "" {
		if sec, err := strconv.ParseInt(v, 10, 64); err == nil {
			req.DateFrom = timestamppb.New(time.Unix(sec, 0).UTC())
		}
	}
	if v := c.Query("date_to"); v != "" {
		if sec, err := strconv.ParseInt(v, 10, 64); err == nil {
			req.DateTo = timestamppb.New(time.Unix(sec, 0).UTC())
		}
	}

	// defaults
	if req.Page < 1 {
		req.Page = 1
	}
	if req.Limit <= 0 {
		req.Limit = 10
	}

	resp, err := s.Usecase.GetQuestionnairesList(c, req)
	if err != nil {
		s.log.Error("get questionnaires failed", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"message": "Некорректные данные"})
		return
	}
	b, err := marshalJSON.Marshal(resp)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "marshal error"})
		return
	}
	c.Data(http.StatusOK, "application/json", b)
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
