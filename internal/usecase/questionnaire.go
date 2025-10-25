package usecase

import (
	"admin_history/internal/entities"
	"admin_history/internal/misc"
	protos "admin_history/pkg/proto/gen/go"
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strings"

	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

// /2025-10-22T20:12:00.340+0300    ERROR   usecase/video.go:118    send final media failed {"error": "call telegram sendPhoto: Post \"https://api.telegram.org/bot7858517388:AAEoU0Or3bii3yfv4pmQR7d2Oxl8_AJZBkA/sendPhoto\": net/http: TLS handshake timeout", "user_id": 686171972, "questionnaire_id": 1}
const demoCaption = `Вот демо-версия вашей истории 🎬✨

Если нравится — оплатите и получите полную версию без водяных знаков + бонус: все фото в стиле Disney!

📲 Больше примеров и идей:
VK: https://vk.com/istoriym
Instagram: https://instagram.com/istoriym_bot`

func (u *Usecase) GetQuestionnairesList(ctx context.Context, req *protos.QuestionnairesListRequest) (*protos.QuestionnairesListResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("nil request")
	}
	if req.Page < 1 {
		req.Page = 1
	}
	if req.Limit <= 0 {
		req.Limit = 10
	}

	var f entities.QuestionnaireFilter
	if req.Payment != nil {
		v := req.Payment.Value
		f.Payment = &v
	}
	if req.Status != nil {
		v := req.Status.Value
		f.Status = &v
	}
	if req.DateFrom != nil {
		t := req.DateFrom.AsTime().UTC()
		f.DateFrom = &t
	}
	if req.DateTo != nil {
		t := req.DateTo.AsTime().UTC()
		f.DateTo = &t
	}
	if req.UserId != nil {
		v := req.UserId.Value
		f.UserID = &v
	}

	items, err := u.questionnaires.GetQuestionnairesList(ctx, req.Page, req.Limit, f)
	if err != nil {
		return nil, err
	}
	count, err := u.questionnaires.CountQuestionnaires(ctx, f)
	resp := &protos.QuestionnairesListResponse{
		Questionnaires: make([]*protos.Questionnaire, 0, len(items)),
		Total:          count,
	}
	for i := range items {
		resp.Questionnaires = append(resp.Questionnaires, items[i].ToProto())
	}
	return resp, nil
}

func (u *Usecase) GetQuestionnaire(ctx context.Context, req *protos.QuestionnaireRequest) (*protos.QuestionnaireResponse, error) {
	if req == nil || req.Id <= 0 {
		return nil, fmt.Errorf("bad id")
	}
	q, err := u.questionnaires.GetQuestionnaire(ctx, &entities.Questionnaire{ID: req.Id})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("user not found")
		}
		return nil, err
	}
	return &protos.QuestionnaireResponse{Questionnaire: q.ToProto()}, nil
}

func (u *Usecase) UpdateQuestionnaire(ctx context.Context, req *protos.UpdateQuestionnaireRequest) (*protos.Status, error) {
	if req == nil || req.Questionnaire == nil || req.Questionnaire.Id == 0 {
		return nil, fmt.Errorf("invalid request")
	}

	ent, err := misc.BuildEntityForUpdate(req.Questionnaire)
	if err != nil {
		return nil, err
	}

	if err := u.questionnaires.UpdateQuestionnaire(ctx, ent); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return &protos.Status{Ok: false, Message: "questionnaire not found"}, nil
		}
		return nil, err
	}

	return &protos.Status{Ok: true, Message: "updated"}, nil
}

func (u *Usecase) SubmitQuestionnaireMedia(ctx context.Context, req *protos.SubmitQuestionnaireMediaRequest, media *MediaUpload) (*protos.Status, error) {
	if req == nil {
		return nil, fmt.Errorf("invalid request")
	}
	if req.QuestionnaireId <= 0 {
		return nil, fmt.Errorf("invalid questionnaire_id")
	}
	if req.UserId <= 0 {
		return nil, fmt.Errorf("invalid user_id")
	}

	qID := req.QuestionnaireId
	mediaAdded := false

	if media != nil {
		demoPhotoPaths, err := u.savePhotoFiles(ctx, media.DemoPhotos)
		if err != nil {
			return nil, fmt.Errorf("save demo photos: %w", err)
		}
		if len(demoPhotoPaths) > 0 {
			req.DemoPhotos = append(req.DemoPhotos, demoPhotoPaths...)
		}

		finalPhotoPaths, err := u.savePhotoFiles(ctx, media.FinalPhotos)
		if err != nil {
			return nil, fmt.Errorf("save final photos: %w", err)
		}
		if len(finalPhotoPaths) > 0 {
			for _, path := range finalPhotoPaths {
				scene := u.nextFinalScene(media)
				req.FinalPhotos = append(req.FinalPhotos, &protos.Photo{
					Path:            path,
					QuestionnaireId: qID,
					Scene:           scene,
					TypePhoto:       "send",
				})
			}
		}

		if media.DeliveryPhoto != nil {
			deliveryPath, err := u.savePhotoFile(ctx, media.DeliveryPhoto)
			if err != nil {
				return nil, fmt.Errorf("save delivery photo: %w", err)
			}
			if deliveryPath != "" {
				req.FinalPhotos = append(req.FinalPhotos, &protos.Photo{
					Path:            deliveryPath,
					QuestionnaireId: qID,
					Scene:           "delivery",
					TypePhoto:       "send",
				})
			}
		}

		if media.DemoVideo != nil {
			p, err := u.saveVideoFile(ctx, media.DemoVideo)
			if err != nil {
				return nil, fmt.Errorf("save demo video: %w", err)
			}
			if p != "" {
				req.DemoVideo = p
			}
		}

		if media.GeneratedVideo != nil {
			p, err := u.saveVideoFile(ctx, media.GeneratedVideo)
			if err != nil {
				return nil, fmt.Errorf("save generated video: %w", err)
			}
			if p != "" {
				req.GeneratedVideo = &protos.Video{
					Path:            p,
					QuestionnaireId: qID,
					TypeVideo:       "send",
				}
			}
		}
	}

	demoPhotoSet := make(map[string]struct{}, len(req.DemoPhotos))
	demoPhotos := make([]string, 0, len(req.DemoPhotos))
	for _, rel := range req.DemoPhotos {
		if abs := resolveStoragePath(rel); abs != "" {
			if _, exists := demoPhotoSet[abs]; exists {
				continue
			}
			demoPhotoSet[abs] = struct{}{}
			demoPhotos = append(demoPhotos, abs)
		}
	}

	demoVideoPath := resolveStoragePath(req.GetDemoVideo())

	if u.photos != nil {
		photos, err := u.photos.GetPhotosQuestionnaire(ctx, qID, "demo")
		if err != nil {
			u.log.Warn("load demo photos failed", zap.Error(err), zap.Int64("questionnaire_id", qID))
		} else {
			for _, photo := range photos {
				if abs := resolveStoragePath(photo.Path); abs != "" {
					if _, exists := demoPhotoSet[abs]; exists {
						continue
					}
					demoPhotoSet[abs] = struct{}{}
					demoPhotos = append(demoPhotos, abs)
				}
			}
		}
	}

	if u.tg != nil && (len(demoPhotos) > 0 || demoVideoPath != "") {
		if err := u.tg.SendDemoMedia(ctx, req.UserId, qID, demoPhotos, demoVideoPath, demoCaption); err != nil {
			u.log.Error("send demo media failed", zap.Error(err), zap.Int64("user_id", req.UserId), zap.Int64("questionnaire_id", qID))
			return nil, err
		}
	}

	for _, path := range req.DemoPhotos {
		rel := strings.TrimSpace(path)
		if rel == "" {
			continue
		}
		ent := entities.Photo{
			QuestionnaireID: qID,
			Path:            rel,
			Scene:           "demo",
			TypePhoto:       "demo",
		}
		if err := u.photos.UploadPhoto(ctx, &ent); err != nil {
			return nil, err
		}
		mediaAdded = true
	}

	for _, p := range req.FinalPhotos {
		if p == nil {
			continue
		}
		path := strings.TrimSpace(p.GetPath())
		if path == "" {
			continue
		}

		scene := strings.TrimSpace(p.GetScene())
		if scene == "" {
			scene = "result"
		}
		typePhoto := strings.TrimSpace(p.GetTypePhoto())
		if typePhoto == "" || !isAllowedPhotoType(typePhoto) {
			typePhoto = "send"
		}
		ent := entities.Photo{
			QuestionnaireID: qID,
			Path:            path,
			Scene:           scene,
			TypePhoto:       typePhoto,
		}
		if err := u.photos.UploadPhoto(ctx, &ent); err != nil {
			return nil, err
		}
		mediaAdded = true
	}

	if video := req.GetGeneratedVideo(); video != nil {
		path := strings.TrimSpace(video.GetPath())
		if path != "" {
			typeVideo := strings.TrimSpace(video.GetTypeVideo())
			if typeVideo == "" || !isAllowedVideoType(typeVideo) {
				typeVideo = "send"
			}
			ent := entities.Video{
				QuestionnaireID: qID,
				Path:            path,
				TypeVideo:       typeVideo,
			}
			if err := u.videos.UploadVideo(ctx, &ent); err != nil {
				return nil, err
			}
			mediaAdded = true
		}
	}

	if demo := strings.TrimSpace(req.GetDemoVideo()); demo != "" {
		ent := entities.Video{
			QuestionnaireID: qID,
			Path:            demo,
			TypeVideo:       "demo",
		}
		if err := u.videos.UploadVideo(ctx, &ent); err != nil {
			return nil, err
		}
		mediaAdded = true
	}

	if mediaAdded {
		if err := u.questionnaires.SetQuestionnaireStatus(ctx, qID, true); err != nil {
			return nil, err
		}
	}

	return &protos.Status{Ok: true, Message: "saved"}, nil
}

func resolveStoragePath(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return ""
	}
	if filepath.IsAbs(path) {
		return path
	}
	trimmed := strings.TrimLeft(path, "/")
	if trimmed == "" {
		return ""
	}
	cleaned := filepath.Clean(filepath.FromSlash(trimmed))
	if strings.HasPrefix(cleaned, "..") {
		return ""
	}
	return filepath.Join("data", cleaned)
}

func (u *Usecase) savePhotoFiles(ctx context.Context, files []*multipart.FileHeader) ([]string, error) {
	if len(files) == 0 {
		return nil, nil
	}
	if u.st == nil {
		return nil, fmt.Errorf("storage not configured")
	}
	saved := make([]string, 0, len(files))
	for _, hdr := range files {
		if hdr == nil {
			continue
		}
		src, err := hdr.Open()
		if err != nil {
			return nil, fmt.Errorf("open photo: %w", err)
		}
		func() {
			defer src.Close()
			ext := normalizeExt(hdr.Filename)
			rel, saveErr := u.st.Save(ctx, src, ext)
			if saveErr != nil {
				err = fmt.Errorf("save photo: %w", saveErr)
				return
			}
			saved = append(saved, rel)
		}()
		if err != nil {
			return nil, err
		}
	}
	return saved, nil
}

func (u *Usecase) saveVideoFile(ctx context.Context, hdr *multipart.FileHeader) (string, error) {
	if hdr == nil {
		return "", nil
	}
	if u.st == nil {
		return "", fmt.Errorf("storage not configured")
	}
	src, err := hdr.Open()
	if err != nil {
		return "", fmt.Errorf("open video: %w", err)
	}
	defer src.Close()

	ext := normalizeVideoExt(hdr.Filename)
	if ext == ".bin" {
		if ct := hdr.Header.Get("Content-Type"); ct != "" {
			if alt := normalizeVideoExt(ct); alt != ".bin" {
				ext = alt
			}
		}
	}
	rel, err := u.st.SaveTo(ctx, "videos", src, ext)
	if err != nil {
		return "", fmt.Errorf("save video: %w", err)
	}
	return rel, nil
}

func (u *Usecase) nextFinalScene(media *MediaUpload) string {
	if media == nil || len(media.FinalPhotoScenes) == 0 {
		return "send"
	}
	scene := strings.TrimSpace(media.FinalPhotoScenes[0])
	media.FinalPhotoScenes = media.FinalPhotoScenes[1:]
	if scene == "" {
		return "send"
	}
	return scene
}

func isAllowedPhotoType(t string) bool {
	t = strings.ToLower(strings.TrimSpace(t))
	switch t {
	case "original", "generated", "send", "demo":
		return true
	default:
		return false
	}
}

func isAllowedVideoType(t string) bool {
	t = strings.ToLower(strings.TrimSpace(t))
	switch t {
	case "send", "demo":
		return true
	default:
		return false
	}
}

func (u *Usecase) savePhotoFile(ctx context.Context, hdr *multipart.FileHeader) (string, error) {
	paths, err := u.savePhotoFiles(ctx, []*multipart.FileHeader{hdr})
	if err != nil {
		return "", err
	}
	if len(paths) == 0 {
		return "", nil
	}
	return paths[0], nil
}
