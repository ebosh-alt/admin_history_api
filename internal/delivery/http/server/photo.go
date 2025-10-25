package server

import (
	protos "admin_history/pkg/proto/gen/go"
	"mime/multipart"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

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
func (s *Server) GetPhotosQuestionnaire(c *gin.Context) {
	req := &protos.PhotoRequest{}
	if v := c.Query("questionnaire_id"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 32); err == nil {
			req.QuestionnaireId = n
		}
	}
	if v := c.Query("type"); v != "" {
		req.Type = v
	}
	//if err := c.ShouldBindJSON(&req); err != nil {
	//	s.log.Error("failed to get questionnaire: ", zap.Error(err))
	//	c.JSON(http.StatusBadRequest, gin.H{"message": "Не корректный запрос"})
	//	return
	//}

	proto, err := s.Usecase.GetPhotosQuestionnaire(c, req)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Не корректные данные"})
		return
	}
	c.JSON(http.StatusOK, proto)
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
func (s *Server) UploadPhoto(c *gin.Context) {
	if err := c.Request.ParseMultipartForm(20 << 20); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid multipart"})
		return
	}

	hdr, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "file required"})
		return
	}

	src, err := hdr.Open() // получаем io.ReadCloser из FileHeader
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "cannot open file"})
		return
	}
	defer func(src multipart.File) {
		err := src.Close()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "error closing file"})
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
	resp, err := s.Usecase.UploadPhoto(c, src, hdr.Filename, photoProto)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	b, _ := marshalJSON.Marshal(resp)
	c.Data(http.StatusOK, "application/json", b)
}

var _ InterfacePhotoServer = (*Server)(nil)
