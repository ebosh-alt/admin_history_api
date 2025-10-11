package server

import (
	protos "admin_history/pkg/proto/gen/go"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

// GetReview godoc
// @Summary      Получить отзыв
// @Description  Возвращает отзыв по идентификатору
// @Tags         reviews
// @Produce      json
// @Param        id   path      int  true  "Review ID"
// @Success      200  {object}  ReviewResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Router       /reviews/{id} [get]
func (s *Server) GetReview(c *gin.Context) {
	req := &protos.ReviewRequest{}

	if v := c.Param("id"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			req.Id = n
		}
	}

	if req.Id <= 0 {
		s.log.Error("invalid review id", zap.Int64("id", req.Id))
		c.JSON(http.StatusBadRequest, gin.H{"message": "Не корректный ID отзыва"})
		return
	}

	reviewProto, err := s.Usecase.GetReview(c, req)
	if err != nil {
		s.log.Error("failed to get review", zap.Error(err), zap.Any("request", req))
		c.JSON(http.StatusNotFound, gin.H{"message": "Отзыв не найден"})
		return
	}

	c.JSON(http.StatusOK, reviewProto)
}

// ReviewsList godoc
// @Summary      Список отзывов
// @Description  Возвращает список отзывов с фильтрами и пагинацией
// @Tags         reviews
// @Produce      json
// @Param        page       query     int     false  "Номер страницы"     minimum(1)
// @Param        limit      query     int     false  "Размер страницы"     minimum(1)
// @Param        user_id    query     int     false  "Фильтр по ID пользователя"
// @Param        date_from  query     string  false  "Дата с (unix или YYYY-MM-DD)"
// @Param        date_to    query     string  false  "Дата по (unix или YYYY-MM-DD)"
// @Success      200        {object}  ReviewsListResponse
// @Failure      400        {object}  ErrorResponse
// @Router       /reviews [get]
func (s *Server) ReviewsList(c *gin.Context) {
	req := &protos.ReviewsListRequest{}

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

	// user_id filter
	if v := c.Query("user_id"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil && n > 0 {
			req.UserId = wrapperspb.Int64(n)
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

	// set defaults
	if req.Page < 1 {
		req.Page = 1
	}
	if req.Limit <= 0 {
		req.Limit = 50
	}

	resp, err := s.Usecase.ReviewsList(c, req)
	if err != nil {
		s.log.Error("failed to get reviews list", zap.Error(err), zap.Any("request", req))
		c.JSON(http.StatusBadRequest, gin.H{"message": "Не корректные данные"})
		return
	}

	b, _ := marshalJSON.Marshal(resp)
	c.Data(http.StatusOK, "application/json", b)
}

var _ InterfaceReviewServer = (*Server)(nil)
