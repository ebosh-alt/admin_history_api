package questionnaire

import (
	"admin_history/internal/entities"
	"context"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	protos "admin_history/pkg/proto/gen/go"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type Usecase interface {
	GetQuestionnairesList(ctx context.Context, req *protos.QuestionnairesListRequest) (*protos.QuestionnairesListResponse, error)
	GetQuestionnaire(ctx context.Context, req *protos.QuestionnaireRequest) (*protos.QuestionnaireResponse, error)
	UpdateQuestionnaire(ctx context.Context, req *protos.UpdateQuestionnaireRequest) (*protos.Status, error)
	SubmitQuestionnaireMedia(ctx context.Context, req *protos.SubmitQuestionnaireMediaRequest, media *entities.MediaUpload) (*protos.Status, error)
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
	r.GET("/:id", h.GetQuestionnaire)
	r.GET("", h.QuestionnairesList)
	r.POST("/update", h.UpdateQuestionnaire)
	r.POST("/media", h.SubmitQuestionnaireMedia)
}

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
func (h *Handler) GetQuestionnaire(c *gin.Context) {
	req := &protos.QuestionnaireRequest{}

	if v := c.Param("id"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 32); err == nil {
			req.Id = n
		}
	}

	qProto, err := h.uc.GetQuestionnaire(c, req)
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
// @Param        user_id    query     int     false  "Фильтр по пользователю"
// @Success      200        {object}  QuestionnairesListResponse
// @Failure      400        {object}  ErrorResponse
// @Failure      404        {object}  ErrorResponse
// @Router       /questionnaires [get]
func (h *Handler) QuestionnairesList(c *gin.Context) {
	req := &protos.QuestionnairesListRequest{}

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

	if v := c.Query("user_id"); v != "" {
		if id, err := strconv.ParseInt(v, 10, 64); err == nil && id > 0 {
			req.UserId = wrapperspb.Int64(id)
		}
	}

	moscowTZ, _ := time.LoadLocation("Europe/Moscow")

	if v := c.Query("date_from"); v != "" {
		var t time.Time
		var err error
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

	resp, err := h.uc.GetQuestionnairesList(c, req)
	if err != nil {
		h.log.Error("list questionnaires failed", zap.Error(err), zap.Any("req", req))
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
func (h *Handler) UpdateQuestionnaire(c *gin.Context) {
	body, _ := io.ReadAll(c.Request.Body)
	var req protos.UpdateQuestionnaireRequest
	if err := unmarshalJSON.Unmarshal(body, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid JSON"})
		return
	}

	resp, err := h.uc.UpdateQuestionnaire(c, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	b, _ := marshalJSON.Marshal(resp)
	c.Data(http.StatusOK, "application/json", b)
}

// SubmitQuestionnaireMedia godoc
// @Summary      Сохранить медиа анкеты
// @Description  Сохраняет медиа анкеты. Поддерживает multipart/form-data (для загрузки файлов) и JSON (тело SubmitQuestionnaireMediaRequest)
// @Tags         questionnaires
// @Accept       multipart/form-data
// @Produce      json
// @Param        questionnaire_id       formData  int     true   "ID анкеты"
// @Param        user_id                formData  int     true   "ID пользователя (chat_id в Telegram)"
// @Param        demo_photos            formData  file    false  "Демо-фотографии (можно несколько файлов)"
// @Param        final_photos           formData  file    false  "Финальные фото (можно несколько файлов)"
// @Param        demo_video             formData  file    false  "Демо-видео"
// @Param        generated_video        formData  file    false  "Сгенерированное видео"
// @Param        delivery_photo         formData  file    false  "Фото для отправки пользователю"
// @Param        demo_photo_path        formData  string  false  "Путь к уже сохранённому демо-фото"
// @Param        final_photo_path       formData  string  false  "Путь к уже сохранённому финальному фото"
// @Param        delivery_photo_path    formData  string  false  "Путь к фото для отправки пользователю"
// @Param        demo_video_path        formData  string  false  "Путь к уже сохранённому демо-видео"
// @Param        generated_video_path   formData  string  false  "Путь к уже сохранённому сгенерированному видео"
// @Param        final_photo_scene      formData  string  false  "Подписи сцен для финальных фото (по порядку файлов)"
// @Param        payload                formData  string  false  "Дополнительный JSON (SubmitQuestionnaireMediaRequest)"
// @Success      200      {object}  Status
// @Failure      400      {object}  ErrorResponse
// @Failure      500      {object}  ErrorResponse
// @Router       /questionnaires/media [post]
func (h *Handler) SubmitQuestionnaireMedia(c *gin.Context) {
	if strings.HasPrefix(c.GetHeader("Content-Type"), "multipart/form-data") {
		h.submitQuestionnaireMediaMultipart(c)
		return
	}

	body, _ := io.ReadAll(c.Request.Body)
	var req protos.SubmitQuestionnaireMediaRequest
	if err := unmarshalJSON.Unmarshal(body, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid JSON"})
		return
	}

	resp, err := h.uc.SubmitQuestionnaireMedia(c, &req, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	b, _ := marshalJSON.Marshal(resp)
	c.Data(http.StatusOK, "application/json", b)
}

func (h *Handler) submitQuestionnaireMediaMultipart(c *gin.Context) {
	if err := c.Request.ParseMultipartForm(500 << 20); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid multipart"})
		return
	}

	var req protos.SubmitQuestionnaireMediaRequest
	if payload := c.PostForm("payload"); payload != "" {
		if err := unmarshalJSON.Unmarshal([]byte(payload), &req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message": "invalid payload json"})
			return
		}
	}

	qid, err := strconv.ParseInt(c.PostForm("questionnaire_id"), 10, 64)
	if err != nil || qid <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid questionnaire_id"})
		return
	}
	req.QuestionnaireId = qid

	uid, err := strconv.ParseInt(c.PostForm("user_id"), 10, 64)
	if err != nil || uid <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid user_id"})
		return
	}
	req.UserId = uid

	finalSceneQueue := make([]string, 0)
	if len(req.FinalPhotos) > 0 {
		prefilled := make([]*protos.Photo, 0, len(req.FinalPhotos))
		for _, meta := range req.FinalPhotos {
			if meta == nil {
				continue
			}
			scene := strings.TrimSpace(meta.Scene)
			if scene == "" {
				scene = "result"
			}
			typePhoto := strings.TrimSpace(meta.TypePhoto)
			if typePhoto == "" {
				typePhoto = "result"
			}
			path := strings.TrimSpace(meta.Path)
			if path != "" {
				prefilled = append(prefilled, &protos.Photo{
					Path:            path,
					QuestionnaireId: qid,
					Scene:           scene,
					TypePhoto:       typePhoto,
				})
			} else {
				finalSceneQueue = append(finalSceneQueue, scene)
			}
		}
		req.FinalPhotos = prefilled
	} else {
		req.FinalPhotos = nil
	}

	for _, scene := range c.PostFormArray("final_photo_scene") {
		if sc := strings.TrimSpace(scene); sc != "" {
			finalSceneQueue = append(finalSceneQueue, sc)
		}
	}

	for _, path := range c.PostFormArray("demo_photo_path") {
		if p := strings.TrimSpace(path); p != "" {
			req.DemoPhotos = append(req.DemoPhotos, p)
		}
	}

	for _, path := range c.PostFormArray("final_photo_path") {
		if p := strings.TrimSpace(path); p != "" {
			req.FinalPhotos = append(req.FinalPhotos, &protos.Photo{
				Path:            p,
				QuestionnaireId: qid,
				Scene:           popScene(&finalSceneQueue),
				TypePhoto:       "send",
			})
		}
	}

	if v := strings.TrimSpace(c.PostForm("delivery_photo_path")); v != "" {
		req.FinalPhotos = append(req.FinalPhotos, &protos.Photo{
			Path:            v,
			QuestionnaireId: qid,
			Scene:           "delivery",
			TypePhoto:       "send",
		})
	}

	if v := strings.TrimSpace(c.PostForm("demo_video_path")); v != "" {
		req.DemoVideo = v
	}
	if v := strings.TrimSpace(c.PostForm("generated_video_path")); v != "" {
		req.GeneratedVideo = &protos.Video{
			Path:            v,
			QuestionnaireId: qid,
			TypeVideo:       "generated",
		}
	}

	media := &entities.MediaUpload{
		FinalPhotoScenes: finalSceneQueue,
	}

	if form := c.Request.MultipartForm; form != nil {
		if files := form.File["demo_photos"]; len(files) > 0 {
			media.DemoPhotos = files
		}
		if files := form.File["final_photos"]; len(files) > 0 {
			media.FinalPhotos = files
		}
		if files := form.File["demo_video"]; len(files) > 0 {
			media.DemoVideo = files[0]
		}
		if files := form.File["generated_video"]; len(files) > 0 {
			media.GeneratedVideo = files[0]
		}
		if files := form.File["delivery_photo"]; len(files) > 0 {
			media.DeliveryPhoto = files[0]
		}
	}

	if media.DemoPhotos == nil && media.FinalPhotos == nil && media.DemoVideo == nil && media.GeneratedVideo == nil && media.DeliveryPhoto == nil && len(media.FinalPhotoScenes) == 0 {
		media = nil
	}

	resp, err := h.uc.SubmitQuestionnaireMedia(c, &req, media)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	b, _ := marshalJSON.Marshal(resp)
	c.Data(http.StatusOK, "application/json", b)
}

func popScene(queue *[]string) string {
	if queue == nil || len(*queue) == 0 {
		return "send"
	}
	scene := strings.TrimSpace((*queue)[0])
	*queue = (*queue)[1:]
	if scene == "" {
		return "send"
	}
	return scene
}
