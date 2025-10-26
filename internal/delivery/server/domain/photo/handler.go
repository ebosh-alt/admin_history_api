package photo

import (
	"context"
	"google.golang.org/protobuf/encoding/protojson"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"

	protos "admin_history/pkg/proto/gen/go"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var (
	unmarshalJSON = protojson.UnmarshalOptions{DiscardUnknown: true}
	marshalJSON   = protojson.MarshalOptions{EmitUnpopulated: true}
)

type Usecase interface {
	GetPhotosQuestionnaire(ctx context.Context, req *protos.PhotoRequest) (*protos.PhotoResponse, error)
	UploadPhoto(ctx context.Context, r io.Reader, filename string, photo *protos.Photo) (*protos.Status, error)
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
	r.GET("", h.GetPhotosQuestionnaire)
	r.POST("/upload", h.UploadPhoto)
}

// GetPhotosQuestionnaire godoc
// @Summary      Получить фото анкеты
// @Description  Возвращает список фотографий анкеты
// @Tags         photos
// @Produce      json
// @Param        questionnaire_id  query     int     true   "ID анкеты"
// @Param        type              query     string  false  "Тип фотографии (original по умолчанию)"
// @Success      200               {object}  PhotoResponse
// @Failure      404               {object}  ErrorResponse
// @Router       /photos [get]
func (h *Handler) GetPhotosQuestionnaire(c *gin.Context) {
	req := &protos.PhotoRequest{}
	if v := c.Query("questionnaire_id"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 32); err == nil {
			req.QuestionnaireId = n
		}
	}
	if v := c.Query("type"); v != "" {
		req.Type = v
	}

	resp, err := h.uc.GetPhotosQuestionnaire(c, req)
	if err != nil {
		h.log.Error("failed to get photos questionnaire", zap.Error(err), zap.Any("request", req))
		c.JSON(http.StatusNotFound, gin.H{"message": "Не корректные данные"})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// UploadPhoto godoc
// @Summary      Загрузить фото
// @Description  Загружает фотографию для анкеты
// @Tags         photos
// @Accept       multipart/form-data
// @Produce      json
// @Param        questionnaire_id  formData  int     true   "ID анкеты"
// @Param        scene             formData  string  false  "Сцена"
// @Param        type              formData  string  false  "Тип фотографии"
// @Param        file              formData  file    true   "Файл изображения"
// @Success      200               {object}  Status
// @Failure      400               {object}  ErrorResponse
// @Router       /photos/upload [post]
func (h *Handler) UploadPhoto(c *gin.Context) {
	if err := c.Request.ParseMultipartForm(20 << 20); err != nil {
		h.log.Error("invalid multipart form", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid multipart"})
		return
	}

	hdr, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "file required"})
		return
	}

	src, err := hdr.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "cannot open file"})
		return
	}
	defer func(src multipart.File) {
		if err := src.Close(); err != nil {
			h.log.Warn("error closing file", zap.Error(err))
		}
	}(src)

	qid, _ := strconv.ParseInt(c.PostForm("questionnaire_id"), 10, 64)
	scene := c.PostForm("scene")
	typePhoto := c.PostForm("type")
	photoProto := &protos.Photo{
		QuestionnaireId: qid,
		Scene:           scene,
		TypePhoto:       typePhoto,
	}

	resp, err := h.uc.UploadPhoto(c, src, hdr.Filename, photoProto)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	b, _ := marshalJSON.Marshal(resp)
	c.Data(http.StatusOK, "application/json", b)
}
