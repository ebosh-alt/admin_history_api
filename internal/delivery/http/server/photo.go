package server

import (
	protos "admin_history/pkg/proto/gen/go"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"mime/multipart"
	"net/http"
	"strconv"
)

func (s *Server) GetPhotosQuestionnaire(c *gin.Context) {
	req := &protos.PhotoRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		s.log.Error("failed to get questionnaire: ", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"message": "Не корректный запрос"})
		return
	}

	proto, err := s.Usecase.GetPhotosQuestionnaire(c, req)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Не корректные данные"})
		return
	}
	c.JSON(http.StatusOK, proto)
}

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
