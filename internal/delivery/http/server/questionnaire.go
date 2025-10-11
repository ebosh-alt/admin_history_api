package server

import (
	protos "admin_history/pkg/proto/gen/go"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

var (
	unmarshalJSON = protojson.UnmarshalOptions{DiscardUnknown: true}
	marshalJSON   = protojson.MarshalOptions{EmitUnpopulated: true}
)

// GetQuestionnaire godoc
// @Summary      Получить анкету
// @Description  Возвращает анкету по идентификатору
// @Tags         questionnaires
// @Produce      json
// @Param        id   path      int  true  "Questionnaire ID"
// @Success      200  {object}  QuestionnaireResponse
// @Failure      404  {object}  ErrorResponse
// @Router       /questionnaires/{id} [get]
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

// QuestionnairesList godoc
// @Summary      Список анкет
// @Description  Возвращает список анкет с фильтрами и пагинацией
// @Tags         questionnaires
// @Produce      json
// @Param        page       query     int     false  "Номер страницы"     minimum(1)
// @Param        limit      query     int     false  "Размер страницы"     minimum(1)
// @Param        payment    query     bool    false  "Фильтр по оплате"
// @Param        status     query     bool    false  "Фильтр по статусу"
// @Param        date_from  query     string  false  "Дата с (unix или YYYY-MM-DD)"
// @Param        date_to    query     string  false  "Дата по (unix или YYYY-MM-DD)"
// @Success      200        {object}  QuestionnairesListResponse
// @Failure      400        {object}  ErrorResponse
// @Failure      404        {object}  ErrorResponse
// @Router       /questionnaires [get]
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

	// UTC+3 timezone for date parsing
	moscowTZ, _ := time.LoadLocation("Europe/Moscow")

	// Parse date_from - support both Unix timestamp and ISO date format
	if v := c.Query("date_from"); v != "" {
		var t time.Time
		var err error

		// Try to parse as Unix timestamp first
		if sec, parseErr := strconv.ParseInt(v, 10, 64); parseErr == nil && sec > 0 {
			t = time.Unix(sec, 0)
		} else {
			// Try to parse as ISO date format (YYYY-MM-DD)
			if t, err = time.ParseInLocation("2006-01-02", v, moscowTZ); err != nil {
				// Try to parse as ISO datetime format (YYYY-MM-DDTHH:MM:SS)
				if t, err = time.ParseInLocation("2006-01-02T15:04:05", v, moscowTZ); err != nil {
					s.log.Error("failed to parse date_from", zap.String("value", v), zap.Error(err))
					c.JSON(http.StatusBadRequest, gin.H{"message": "Не корректный формат даты date_from"})
					return
				}
			}
		}

		// Convert to UTC for database storage
		req.DateFrom = timestamppb.New(t.UTC())
	}

	// Parse date_to - support both Unix timestamp and ISO date format
	if v := c.Query("date_to"); v != "" {
		var t time.Time
		var err error

		// Try to parse as Unix timestamp first
		if sec, parseErr := strconv.ParseInt(v, 10, 64); parseErr == nil && sec > 0 {
			t = time.Unix(sec, 0)
		} else {
			// Try to parse as ISO date format (YYYY-MM-DD)
			if t, err = time.ParseInLocation("2006-01-02", v, moscowTZ); err != nil {
				// Try to parse as ISO datetime format (YYYY-MM-DDTHH:MM:SS)
				if t, err = time.ParseInLocation("2006-01-02T15:04:05", v, moscowTZ); err != nil {
					s.log.Error("failed to parse date_to", zap.String("value", v), zap.Error(err))
					c.JSON(http.StatusBadRequest, gin.H{"message": "Не корректный формат даты date_to"})
					return
				}
			}
		}

		// For date_to, if it's just a date (no time), add 23:59:59 to include the whole day
		if t.Hour() == 0 && t.Minute() == 0 && t.Second() == 0 {
			t = t.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
		}

		// Convert to UTC for database storage
		req.DateTo = timestamppb.New(t.UTC())
	}
	resp, err := s.Usecase.GetQuestionnairesList(c, req)
	if err != nil {
		s.log.Error("list questionnaires failed", zap.Error(err), zap.Any("req", req))
		c.JSON(http.StatusNotFound, gin.H{"message": "Не корректные данные"})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// UpdateQuestionnaire godoc
// @Summary      Обновить анкету
// @Description  Обновляет данные анкеты
// @Tags         questionnaires
// @Accept       json
// @Produce      json
// @Param        request  body      UpdateQuestionnaireRequest  true  "Данные анкеты"
// @Success      200      {object}  Status
// @Failure      400      {object}  ErrorResponse
// @Failure      500      {object}  ErrorResponse
// @Router       /questionnaires/update [post]
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

// SubmitQuestionnaireMedia godoc
// @Summary      Сохранить итоговые медиа анкеты
// @Description  Сохраняет финальные фотографии и сгенерированное видео анкеты, а также принимает демо-материалы без сохранения
// @Tags         questionnaires
// @Accept       json
// @Produce      json
// @Param        request  body      SubmitQuestionnaireMediaRequest  true  "Медиа анкеты"
// @Success      200      {object}  Status
// @Failure      400      {object}  ErrorResponse
// @Failure      500      {object}  ErrorResponse
// @Router       /questionnaires/media [post]
func (s *Server) SubmitQuestionnaireMedia(c *gin.Context) {
	body, _ := io.ReadAll(c.Request.Body)
	var req protos.SubmitQuestionnaireMediaRequest
	if err := unmarshalJSON.Unmarshal(body, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid JSON"})
		return
	}

	resp, err := s.Usecase.SubmitQuestionnaireMedia(c, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	b, _ := marshalJSON.Marshal(resp)
	c.Data(http.StatusOK, "application/json", b)
}

var _ InterfaceQuestionnaireServer = (*Server)(nil)
