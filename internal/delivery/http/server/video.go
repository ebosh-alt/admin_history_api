package server

import (
	protos "admin_history/pkg/proto/gen/go"
	"github.com/gin-gonic/gin"
	"mime/multipart"
	"net/http"
	"strconv"
)

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
func (s *Server) GetVideosQuestionnaire(c *gin.Context) {
	req := &protos.VideoRequest{}
	if v := c.Query("questionnaire_id"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 32); err == nil {
			req.QuestionnaireId = n
		}
	}
	if v := c.Query("type"); v != "" {
		req.Type = v
	}

	proto, err := s.Usecase.GetVideosQuestionnaire(c, req)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Не корректные данные"})
		return
	}
	c.JSON(http.StatusOK, proto)
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
func (s *Server) UploadVideo(c *gin.Context) {
	if err := c.Request.ParseMultipartForm(200 << 20); err != nil {
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
			c.JSON(http.StatusBadRequest, gin.H{"message": "error closing file"})
		}
	}(src)

	qid, _ := strconv.ParseInt(c.PostForm("questionnaire_id"), 10, 64)
	typeVideo := c.PostForm("type")

	videoProto := &protos.Video{
		QuestionnaireId: qid,
		TypeVideo:       typeVideo,
	}

	contentType := hdr.Header.Get("Content-Type")
	resp, err := s.Usecase.UploadVideo(c, src, hdr.Filename, contentType, videoProto)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	b, _ := marshalJSON.Marshal(resp)
	c.Data(http.StatusOK, "application/json", b)
}

var _ InterfaceVideoServer = (*Server)(nil)
