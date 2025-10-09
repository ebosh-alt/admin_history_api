package server

import (
	protos "admin_history/pkg/proto/gen/go"
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

	if v := c.Param("id"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 32); err == nil {
			req.Id = n
		}
	}

	qProto, err := s.Usecase.GetQuestionnaire(c, req)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Не корректные данные"})
		return
	}
	c.JSON(http.StatusOK, qProto)
}

// GET /api/questionnaires?payment=true&status=false&date_from=2025-09-29T00:00:00Z&date_to=2025-09-29
func (s *Server) QuestionnairesList(c *gin.Context) {
	req := &protos.QuestionnairesListRequest{}

	// page/limit
	if v := c.Query("page"); v != "" {
		if i, err := strconv.Atoi(v); err == nil && i > 0 {
			req.Page = int32(i)
		}
	}
	if v := c.Query("limit"); v != "" {
		if i, err := strconv.Atoi(v); err == nil && i > 0 {
			req.Limit = int32(i)
		}
	}

	// bools
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

	parseYMD := func(s string) (time.Time, error) {
		t, err := time.ParseInLocation("2006-01-02", s, time.UTC) // бизнес-зона = UTC
		if err != nil {
			return time.Time{}, err
		}
		// начало дня в UTC
		return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC), nil
	}

	var fromPtr, toPtr *time.Time
	if v := c.Query("date_from"); v != "" {
		if t, err := parseYMD(v); err == nil {
			fromPtr = &t
		}
	}
	if v := c.Query("date_to"); v != "" {
		if t, err := parseYMD(v); err == nil {
			toPtr = &t
		}
	}

	if fromPtr != nil && (toPtr == nil || toPtr.Equal(*fromPtr)) {
		t := fromPtr.Add(24 * time.Hour)
		toPtr = &t
	}

	// в proto
	if fromPtr != nil {
		req.DateFrom = timestamppb.New(*fromPtr)
	}
	if toPtr != nil {
		req.DateTo = timestamppb.New(*toPtr)
	}
	resp, err := s.Usecase.GetQuestionnairesList(c, req)
	if err != nil {
		s.log.Error("list questionnaires failed", zap.Error(err), zap.Any("req", req))
		c.JSON(http.StatusNotFound, gin.H{"message": "Не корректные данные"})
		return
	}
	c.JSON(http.StatusOK, resp)
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
