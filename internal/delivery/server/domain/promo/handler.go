package promo

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strconv"

	"admin_history/internal/repository"
	protos "admin_history/pkg/proto/gen/go"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type Usecase interface {
	GetPromoCode(ctx context.Context, req *protos.PromoCodeRequest) (*protos.PromoCodeResponse, error)
	PromoCodesList(ctx context.Context, req *protos.PromoCodesListRequest) (*protos.PromoCodesListResponse, error)
	CreatePromoCode(ctx context.Context, req *protos.CreatePromoCodeRequest) (*protos.Status, error)
	UpdatePromoCode(ctx context.Context, req *protos.UpdatePromoCodeRequest) (*protos.Status, error)
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
	r.GET("/:id", h.GetPromoCode)
	r.GET("", h.PromoCodesList)
	r.POST("", h.CreatePromoCode)
	r.POST("/update", h.UpdatePromoCode)
}

var (
	unmarshalJSON = protojson.UnmarshalOptions{DiscardUnknown: true}
	marshalJSON   = protojson.MarshalOptions{EmitUnpopulated: true}
)

// GetPromoCode godoc
// @Summary      Получить промокод
// @Description  Возвращает информацию о промокоде по идентификатору
// @Tags         promo-codes
// @Produce      json
// @Param        id   path      int  true  "Promo Code ID"
// @Success      200  {object}  PromoCodeResponse
// @Failure      400  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Router       /promo-codes/{id} [get]
func (h *Handler) GetPromoCode(c *gin.Context) {
	req := &protos.PromoCodeRequest{}

	if v := c.Param("id"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			req.Id = n
		}
	}

	if req.Id <= 0 {
		h.log.Error("invalid promo code id", zap.Int64("id", req.Id))
		c.JSON(http.StatusBadRequest, gin.H{"message": "Не корректный ID промокода"})
		return
	}

	promoCodeProto, err := h.uc.GetPromoCode(c, req)
	if err != nil {
		h.log.Error("failed to get promo code", zap.Error(err), zap.Any("request", req))
		c.JSON(http.StatusNotFound, gin.H{"message": "Промокод не найден"})
		return
	}

	c.JSON(http.StatusOK, promoCodeProto)
}

// PromoCodesList godoc
// @Summary      Список промокодов
// @Description  Возвращает список промокодов с фильтрами и пагинацией
// @Tags         promo-codes
// @Produce      json
// @Param        page    query     int   false  "Номер страницы" minimum(1)
// @Param        limit   query     int   false  "Размер страницы" minimum(1)
// @Param        status  query     bool  false  "Фильтр по статусу"
// @Success      200     {object}  PromoCodesListResponse
// @Failure      400     {object}  ErrorResponse
// @Router       /promo-codes [get]
func (h *Handler) PromoCodesList(c *gin.Context) {
	req := &protos.PromoCodesListRequest{}

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

	if v := c.Query("status"); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			req.Status = wrapperspb.Bool(b)
		}
	}

	if req.Page < 1 {
		req.Page = 1
	}
	if req.Limit <= 0 {
		req.Limit = 50
	}

	resp, err := h.uc.PromoCodesList(c, req)
	if err != nil {
		h.log.Error("failed to get promo codes list", zap.Error(err), zap.Any("request", req))
		c.JSON(http.StatusBadRequest, gin.H{"message": "Не корректные данные"})
		return
	}

	b, _ := marshalJSON.Marshal(resp)
	c.Data(http.StatusOK, "application/json", b)
}

// CreatePromoCode godoc
// @Summary      Создать промокод
// @Description  Создает новый промокод
// @Tags         promo-codes
// @Accept       json
// @Produce      json
// @Param        request  body      CreatePromoCodeRequest  true  "Данные промокода"
// @Success      201      {object}  Status
// @Failure      400      {object}  ErrorResponse
// @Failure      500      {object}  ErrorResponse
// @Router       /promo-codes [post]
func (h *Handler) CreatePromoCode(c *gin.Context) {
	req := &protos.CreatePromoCodeRequest{}

	data, err := io.ReadAll(c.Request.Body)
	if err != nil {
		h.log.Error("failed to read request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"message": "Не корректный запрос"})
		return
	}

	if err := unmarshalJSON.Unmarshal(data, req); err != nil {
		h.log.Error("failed to unmarshal request", zap.Error(err), zap.String("request", string(data)))
		c.JSON(http.StatusBadRequest, gin.H{"message": "Не корректный запрос (типы данных или их формат)"})
		return
	}

	if req.PromoCode == nil {
		h.log.Error("promo code is required")
		c.JSON(http.StatusBadRequest, gin.H{"message": "Промокод обязателен"})
		return
	}

	if req.PromoCode.Value == "" {
		h.log.Error("promo code value is required")
		c.JSON(http.StatusBadRequest, gin.H{"message": "Значение промокода обязательно"})
		return
	}

	if req.PromoCode.Percent <= 0 {
		h.log.Error("promo code percent must be positive")
		c.JSON(http.StatusBadRequest, gin.H{"message": "Процент должен быть положительным"})
		return
	}

	if req.PromoCode.Description == "" {
		h.log.Error("promo code description is required")
		c.JSON(http.StatusBadRequest, gin.H{"message": "Описание промокода обязательно"})
		return
	}

	status, err := h.uc.CreatePromoCode(c, req)
	if err != nil {
		h.log.Error("failed to create promo code", zap.Error(err), zap.Any("request", req))
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Ошибка создания промокода"})
		return
	}

	b, _ := marshalJSON.Marshal(status)
	c.Data(http.StatusCreated, "application/json", b)
}

// UpdatePromoCode godoc
// @Summary      Обновить промокод
// @Description  Обновляет существующий промокод
// @Tags         promo-codes
// @Accept       json
// @Produce      json
// @Param        request  body      UpdatePromoCodeRequest  true  "Данные промокода"
// @Success      200      {object}  Status
// @Failure      400      {object}  ErrorResponse
// @Failure      404      {object}  ErrorResponse
// @Failure      500      {object}  ErrorResponse
// @Router       /promo-codes/update [post]
func (h *Handler) UpdatePromoCode(c *gin.Context) {
	req := &protos.UpdatePromoCodeRequest{}

	data, err := io.ReadAll(c.Request.Body)
	if err != nil {
		h.log.Error("failed to read request", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"message": "Не корректный запрос"})
		return
	}

	if err := unmarshalJSON.Unmarshal(data, req); err != nil {
		h.log.Error("failed to unmarshal request", zap.Error(err), zap.String("request", string(data)))
		c.JSON(http.StatusBadRequest, gin.H{"message": "Не корректный запрос (типы данных или их формат)"})
		return
	}

	if req.PromoCode == nil {
		h.log.Error("promo code is required for update")
		c.JSON(http.StatusBadRequest, gin.H{"message": "Промокод обязателен"})
		return
	}

	if req.PromoCode.Id <= 0 {
		h.log.Error("promo code id is required")
		c.JSON(http.StatusBadRequest, gin.H{"message": "Некорректный ID промокода"})
		return
	}

	if req.PromoCode.Value == "" {
		h.log.Error("promo code value is required")
		c.JSON(http.StatusBadRequest, gin.H{"message": "Значение промокода обязательно"})
		return
	}

	if req.PromoCode.Percent <= 0 {
		h.log.Error("promo code percent must be positive")
		c.JSON(http.StatusBadRequest, gin.H{"message": "Процент должен быть положительным"})
		return
	}

	if req.PromoCode.Description == "" {
		h.log.Error("promo code description is required")
		c.JSON(http.StatusBadRequest, gin.H{"message": "Описание промокода обязательно"})
		return
	}

	status, err := h.uc.UpdatePromoCode(c, req)
	if err != nil {
		if errors.Is(err, repository.ErrPromoCodeNotFound) {
			h.log.Warn("promo code not found", zap.Any("request", req))
			c.JSON(http.StatusNotFound, gin.H{"message": "Промокод не найден"})
			return
		}
		h.log.Error("failed to update promo code", zap.Error(err), zap.Any("request", req))
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Ошибка обновления промокода"})
		return
	}

	b, _ := marshalJSON.Marshal(status)
	c.Data(http.StatusOK, "application/json", b)
}
