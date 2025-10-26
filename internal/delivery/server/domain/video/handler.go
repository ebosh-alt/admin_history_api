package video

import (
	"context"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"

	protos "admin_history/pkg/proto/gen/go"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"
)

type Usecase interface {
	GetVideosQuestionnaire(ctx context.Context, req *protos.VideoRequest) (*protos.VideoResponse, error)
	UploadVideo(ctx context.Context, r io.Reader, filename string, contentType string, video *protos.Video) (*protos.Status, error)
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
	r.GET("", h.GetVideosQuestionnaire)
	r.POST("/upload", h.UploadVideo)
}

var marshalJSON = protojson.MarshalOptions{EmitUnpopulated: true}

// GetVideosQuestionnaire godoc
// @Summary      Получить видео анкеты
// @Description  Возвращает список видеозаписей анкеты
// @Tags         videos
// @Produce      json
// @Param        questionnaire_id  query     int     true   "ID анкеты"
// @Param        type              query     string  false  "Тип видео"
// @Success      200               {object}  VideoResponse
// @Failure      404               {object}  ErrorResponse
// @Router       /videos [get]
func (h *Handler) GetVideosQuestionnaire(c *gin.Context) {
	req := &protos.VideoRequest{}
	if v := c.Query("questionnaire_id"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 32); err == nil {
			req.QuestionnaireId = n
		}
	}
	if v := c.Query("type"); v != "" {
		req.Type = v
	}

	resp, err := h.uc.GetVideosQuestionnaire(c, req)
	if err != nil {
		h.log.Error("failed to get videos questionnaire", zap.Error(err), zap.Any("request", req))
		c.JSON(http.StatusNotFound, gin.H{"message": "Не корректные данные"})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// UploadVideo godoc
// @Summary      Загрузить видео
// @Description  Загружает видеозапись для анкеты
// @Tags         videos
// @Accept       multipart/form-data
// @Produce      json
// @Param        questionnaire_id  formData  int     true   "ID анкеты"
// @Param        type              formData  string  false  "Тип видео"
// @Param        file              formData  file    true   "Видео файл"
// @Success      200               {object}  Status
// @Failure      400               {object}  ErrorResponse
// @Router       /videos/upload [post]
func (h *Handler) UploadVideo(c *gin.Context) {
	if err := c.Request.ParseMultipartForm(200 << 20); err != nil {
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
	typeVideo := c.PostForm("type")

	videoProto := &protos.Video{
		QuestionnaireId: qid,
		TypeVideo:       typeVideo,
	}

	contentType := hdr.Header.Get("Content-Type")
	resp, err := h.uc.UploadVideo(c, src, hdr.Filename, contentType, videoProto)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	b, _ := marshalJSON.Marshal(resp)
	c.Data(http.StatusOK, "application/json", b)
}
