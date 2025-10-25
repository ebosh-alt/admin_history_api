package usecase

import (
	"admin_history/internal/entities"
	protos "admin_history/pkg/proto/gen/go"
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"go.uber.org/zap"
)

const finalCaption = `Ваша история готова! ✨

Если потребуется другой формат или правки, просто напишите нам, и мы всё оперативно скорректируем.`

func (u *Usecase) GetVideosQuestionnaire(ctx context.Context, req *protos.VideoRequest) (*protos.VideoResponse, error) {
	qID := req.GetQuestionnaireId()
	typeVideo := strings.TrimSpace(req.GetType())

	videos, err := u.videos.GetVideosQuestionnaire(ctx, qID, typeVideo)
	if err != nil {
		return nil, err
	}

	resp := make([]*protos.Video, 0, len(videos))
	for _, v := range videos {
		resp = append(resp, &protos.Video{
			Path:            v.Path,
			QuestionnaireId: v.QuestionnaireID,
			TypeVideo:       v.TypeVideo,
		})
	}

	return &protos.VideoResponse{Video: resp}, nil
}

func (u *Usecase) UploadVideo(ctx context.Context, r io.Reader, filename string, contentType string, video *protos.Video) (*protos.Status, error) {
	if video.GetQuestionnaireId() <= 0 {
		return nil, fmt.Errorf("invalid questionnaire_id")
	}

	ext := normalizeVideoExt(filename)
	if ext == ".bin" && contentType != "" {
		if ctExt := normalizeVideoExt(contentType); ctExt != ".bin" {
			ext = ctExt
		}
	}
	relPath, err := u.st.SaveTo(ctx, "videos", r, ext)
	if err != nil {
		u.log.Error("storage save failed", zap.Error(err))
		return nil, err
	}
	if relPath == "" {
		return nil, fmt.Errorf("invalid video payload")
	}

	video.Path = relPath

	entity := entities.Video{
		Path:            video.Path,
		QuestionnaireID: video.QuestionnaireId,
		TypeVideo:       normalizeVideoType(video.TypeVideo),
	}

	if err := u.videos.UploadVideo(ctx, &entity); err != nil {
		if dd := u.st.Remove(ctx, relPath); dd != nil {
			u.log.Warn("cleanup remove file failed", zap.String("path", relPath), zap.Error(dd))
		}
		return nil, err
	}

	if err := u.questionnaires.SetQuestionnaireStatus(ctx, video.QuestionnaireId, true); err != nil {
		if dd := u.st.Remove(ctx, relPath); dd != nil {
			u.log.Warn("cleanup remove file failed", zap.String("path", relPath), zap.Error(dd))
		}
		return nil, err
	}

	if u.tg != nil && entity.TypeVideo == "send" {
		q, err := u.questionnaires.GetQuestionnaire(ctx, &entities.Questionnaire{ID: video.QuestionnaireId})
		if err != nil {
			u.log.Error("load questionnaire for final delivery failed", zap.Error(err), zap.Int64("questionnaire_id", video.QuestionnaireId))
			return nil, err
		}
		if q.UserID <= 0 {
			u.log.Warn("questionnaire has no user id for delivery", zap.Int64("questionnaire_id", video.QuestionnaireId))
		} else {
			videoPath := resolveStoragePath(entity.Path)
			if videoPath == "" {
				u.log.Error("resolve final video path failed", zap.String("path", entity.Path), zap.Int64("questionnaire_id", video.QuestionnaireId))
				return nil, fmt.Errorf("invalid video storage path")
			}

			finalPhotoSet := make(map[string]struct{})
			finalPhotoPaths := make([]string, 0)

			if photos, err := u.photos.GetPhotosQuestionnaire(ctx, video.QuestionnaireId, ""); err != nil {
				u.log.Warn("load questionnaire photos failed", zap.Error(err), zap.Int64("questionnaire_id", video.QuestionnaireId))
			} else {
				for _, photo := range photos {
					if !strings.EqualFold(strings.TrimSpace(photo.TypePhoto), "send") {
						continue
					}
					if abs := resolveStoragePath(photo.Path); abs != "" {
						if _, exists := finalPhotoSet[abs]; exists {
							continue
						}
						finalPhotoSet[abs] = struct{}{}
						finalPhotoPaths = append(finalPhotoPaths, abs)
					}
				}
			}

			if err := u.tg.SendFinalMedia(ctx, q.UserID, video.QuestionnaireId, finalPhotoPaths, videoPath, finalCaption); err != nil {
				u.log.Error("send final media failed", zap.Error(err), zap.Int64("user_id", q.UserID), zap.Int64("questionnaire_id", video.QuestionnaireId))
				return nil, err
			}
		}
	}

	return &protos.Status{Ok: true, Message: "created"}, nil
}

func normalizeVideoExt(name string) string {
	name = strings.ToLower(strings.TrimSpace(name))
	if name == "" {
		return ".bin"
	}

	if idx := strings.IndexAny(name, "?#"); idx >= 0 {
		name = strings.TrimSpace(name[:idx])
	}

	if ext := canonicalVideoExt(filepath.Ext(name)); ext != "" {
		return ext
	}

	if !strings.ContainsAny(name, "./") {
		if ext := canonicalVideoExt("." + name); ext != "" {
			return ext
		}
	}

	if strings.Contains(name, "/") {
		parts := strings.Split(name, "/")
		subtype := strings.TrimSpace(parts[len(parts)-1])
		if idx := strings.Index(subtype, ";"); idx >= 0 {
			subtype = strings.TrimSpace(subtype[:idx])
		}
		if subtype != "" {
			if ext := canonicalVideoExt("." + strings.TrimPrefix(subtype, ".")); ext != "" {
				return ext
			}
		}
	}

	return ".bin"
}

func normalizeVideoType(t string) string {
	t = strings.ToLower(strings.TrimSpace(t))
	if isAllowedVideoType(t) {
		return t
	}
	return "send"
}

func canonicalVideoExt(ext string) string {
	ext = strings.ToLower(strings.TrimSpace(ext))
	if ext == "" {
		return ""
	}
	if !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}

	switch ext {
	case ".mp4", ".mov", ".avi", ".mkv", ".webm", ".m4v", ".mpg", ".mpeg", ".3gp":
		return ext
	}
	return ""
}

var _ InterfaceVideoUsecase = (*Usecase)(nil)
