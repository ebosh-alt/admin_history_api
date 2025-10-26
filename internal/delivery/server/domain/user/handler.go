package user

import (
	"context"
	"io"
	"net/http"
	"strconv"
	"time"

	protos "admin_history/pkg/proto/gen/go"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type Usecase interface {
	GetUser(ctx context.Context, req *protos.UserRequest) (*protos.UserResponse, error)
	UsersList(ctx context.Context, req *protos.UsersListRequest) (*protos.UsersListResponse, error)
	UpdateUser(ctx context.Context, req *protos.UpdateUserRequest) (*protos.Status, error)
}

type Handler struct {
	log *zap.Logger
	uc  Usecase
}

func New(log *zap.Logger, uc Usecase) *Handler {
	return &Handler{
		log: log,
		uc:  uc,
	}
}

func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	r.GET("/:id", h.GetUser)
	r.GET("", h.UsersList)
	r.POST("/update", h.UpdateUser)
}

// GetUser godoc
// @Summary      Получить пользователя
// @Description  Возвращает информацию о пользователе по идентификатору
// @Tags         users
// @Produce      json
// @Param        id   path      int  true  "User ID"
// @Success      200  {object}  UserResponse
// @Failure      404  {object}  ErrorResponse
// @Router       /users/{id} [get]
func (h *Handler) GetUser(c *gin.Context) {
	req := &protos.UserRequest{}
	if v := c.Param("id"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 32); err == nil {
			req.Id = n
		}
	}

	userProto, err := h.uc.GetUser(c, req)
	if err != nil {
		h.log.Error("failed to get user: "+err.Error(), zap.Any("request", req))
		c.JSON(http.StatusNotFound, gin.H{"message": "Не корректные данные"})
		return
	}

	m := protojson.MarshalOptions{EmitUnpopulated: true}
	b, err := m.Marshal(userProto)
	if err != nil {
		h.log.Error("failed to marshal user response", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Не удалось сформировать ответ"})
		return
	}
	c.Data(http.StatusOK, "application/json", b)
}

// UsersList godoc
// @Summary      Список пользователей
// @Description  Возвращает список пользователей с фильтрами и пагинацией
// @Tags         users
// @Produce      json
// @Param        page             query     int     false  "Номер страницы"     minimum(1)
// @Param        limit            query     int     false  "Размер страницы"     minimum(1)
// @Param        status           query     bool    false  "Фильтр по статусу"
// @Param        accepted_offer   query     bool    false  "Фильтр по принятию оферты"
// @Param        promocode        query     string  false  "Фильтр по промокоду"
// @Param        age_from         query     int     false  "Возраст от"
// @Param        age_to           query     int     false  "Возраст до"
// @Param        gender           query     string  false  "Пол"
// @Param        map_binding      query     bool    false  "Наличие привязки карты"
// @Param        date_from        query     string  false  "Дата с (unix или YYYY-MM-DD)"
// @Param        date_to          query     string  false  "Дата по (unix или YYYY-MM-DD)"
// @Success      200              {object}  UsersListResponse
// @Failure      400              {object}  ErrorResponse
// @Router       /users [get]
func (h *Handler) UsersList(c *gin.Context) {
	req := &protos.UsersListRequest{}

	if v := c.Query("page"); v != "" {
		if n, _ := strconv.ParseInt(v, 10, 32); n > 0 {
			req.Page = int32(n)
		}
	}
	if v := c.Query("limit"); v != "" {
		if n, _ := strconv.ParseInt(v, 10, 32); n > 0 {
			req.Limit = int32(n)
		}
	}
	if v := c.Query("status"); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			req.Status = wrapperspb.Bool(b)
		}
	}
	if v := c.Query("accepted_offer"); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			req.AcceptedOffer = wrapperspb.Bool(b)
		}
	}
	if v := c.Query("promocode"); v != "" {
		req.Promocode = wrapperspb.String(v)
	}
	if v := c.Query("age_from"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil && n >= 0 {
			req.AgeFrom = wrapperspb.Int64(n)
		}
	}
	if v := c.Query("age_to"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil && n >= 0 {
			req.AgeTo = wrapperspb.Int64(n)
		}
	}
	if v := c.Query("gender"); v != "" {
		req.Gender = wrapperspb.String(v)
	}
	if v := c.Query("map_binding"); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			req.MapBinding = wrapperspb.Bool(b)
		}
	}

	moscowTZ, _ := time.LoadLocation("Europe/Moscow")

	if v := c.Query("date_from"); v != "" {
		var t time.Time
		var err error

		// Try to parse as Unix timestamp first
		if sec, parseErr := strconv.ParseInt(v, 10, 64); parseErr == nil && sec > 0 {
			t = time.Unix(sec, 0)
		} else {
			if t, err = time.ParseInLocation("2006-01-02", v, moscowTZ); err != nil {
				if t, err = time.ParseInLocation("2006-01-02T15:04:05", v, moscowTZ); err != nil {
					h.log.Error("failed to parse date_from", zap.String("value", v), zap.Error(err))
					c.JSON(http.StatusBadRequest, gin.H{"message": "Не корректный формат даты date_from"})
					return
				}
			}
		}

		req.DateFrom = timestamppb.New(t.UTC())
	}

	if v := c.Query("date_to"); v != "" {
		var t time.Time
		var err error

		if sec, parseErr := strconv.ParseInt(v, 10, 64); parseErr == nil && sec > 0 {
			t = time.Unix(sec, 0)
		} else {
			if t, err = time.ParseInLocation("2006-01-02", v, moscowTZ); err != nil {
				if t, err = time.ParseInLocation("2006-01-02T15:04:05", v, moscowTZ); err != nil {
					h.log.Error("failed to parse date_to", zap.String("value", v), zap.Error(err))
					c.JSON(http.StatusBadRequest, gin.H{"message": "Не корректный формат даты date_to"})
					return
				}
			}
		}

		if t.Hour() == 0 && t.Minute() == 0 && t.Second() == 0 {
			t = t.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
		}

		req.DateTo = timestamppb.New(t.UTC())
	}

	if req.Page < 1 {
		req.Page = 1
	}
	if req.Limit <= 0 {
		req.Limit = 50
	}

	resp, err := h.uc.UsersList(c, req)
	if err != nil {
		h.log.Error("failed to get users list", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"message": "Не корректные данные"})
		return
	}

	m := protojson.MarshalOptions{EmitUnpopulated: true}
	b, _ := m.Marshal(resp)
	c.Data(http.StatusOK, "application/json", b)
}

// UpdateUser godoc
// @Summary      Обновить пользователя
// @Description  Обновляет данные пользователя
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        request  body      UpdateUserRequest  true  "Данные пользователя"
// @Success      200      {object}  Status
// @Failure      400      {object}  ErrorResponse
// @Failure      404      {object}  ErrorResponse
// @Router       /users/update [post]
func (h *Handler) UpdateUser(c *gin.Context) {
	req := &protos.UpdateUserRequest{}
	data, err := io.ReadAll(c.Request.Body)
	if err != nil {
		h.log.Error("failed to read request: "+err.Error(), zap.String("request", string(data)))
		c.JSON(http.StatusBadRequest, gin.H{"message": "Не корректный запрос"})
		return
	}
	if err := protojson.Unmarshal(data, req); err != nil {
		h.log.Error("failed to unmarshal request: "+err.Error(), zap.String("request", string(data)))
		c.JSON(http.StatusBadRequest, gin.H{"message": "Не корректный запрос(типы данных или их формат)"})
		return
	}

	status, err := h.uc.UpdateUser(c, req)
	if err != nil {
		h.log.Error("failed to update user: "+err.Error(), zap.Any("request", req))
		c.JSON(http.StatusNotFound, gin.H{"message": "Не корректные данные"})
		return
	}
	c.JSON(http.StatusOK, status)
}
