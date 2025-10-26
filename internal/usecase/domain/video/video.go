package video

import (
	"admin_history/internal/repository"
	"admin_history/internal/storage"
	"admin_history/internal/usecase/domain/base"
	"context"
	"fmt"
	"io"
	"strings"

	"admin_history/internal/entities"
	protos "admin_history/pkg/proto/gen/go"

	"go.uber.org/zap"
)

type Usecase struct {
	log               *zap.Logger
	questionnaireRepo repository.QuestionnaireRepository
	videoRepo         repository.VideoRepository
	photoRepo         repository.PhotoRepository
	st                *storage.FS
}

func New(
	log *zap.Logger,
	questionnaireRepo repository.QuestionnaireRepository,
	videoRepo repository.VideoRepository,
	photoRepo repository.PhotoRepository,
	st *storage.FS,
) *Usecase {
	return &Usecase{
		log:               log,
		questionnaireRepo: questionnaireRepo,
		videoRepo:         videoRepo,
		photoRepo:         photoRepo,
		st:                st,
	}
}

const finalCaption = `Ваша история готова! ✨

Если потребуется другой формат или правки, просто напишите нам, и мы всё оперативно скорректируем.`

func (u *Usecase) GetVideosQuestionnaire(ctx context.Context, req *protos.VideoRequest) (*protos.VideoResponse, error) {
	qID := req.GetQuestionnaireId()
	typeVideo := strings.TrimSpace(req.GetType())

	videos, err := u.videoRepo.GetVideosQuestionnaire(ctx, qID, typeVideo)
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

	ext := base.NormalizeVideoExt(filename)
	if ext == ".bin" && contentType != "" {
		if ctExt := base.NormalizeVideoExt(contentType); ctExt != ".bin" {
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
		TypeVideo:       base.NormalizeVideoType(video.TypeVideo),
	}

	if err := u.videoRepo.UploadVideo(ctx, &entity); err != nil {
		if dd := u.st.Remove(ctx, relPath); dd != nil {
			u.log.Warn("cleanup remove file failed", zap.String("path", relPath), zap.Error(dd))
		}
		return nil, err
	}

	if err := u.questionnaireRepo.SetQuestionnaireStatus(ctx, video.QuestionnaireId, true); err != nil {
		if dd := u.st.Remove(ctx, relPath); dd != nil {
			u.log.Warn("cleanup remove file failed", zap.String("path", relPath), zap.Error(dd))
		}
		return nil, err
	}

	if entity.TypeVideo == "send" {
		q, err := u.questionnaireRepo.GetQuestionnaire(ctx, &entities.Questionnaire{ID: video.QuestionnaireId})
		if err != nil {
			u.log.Error("load questionnaire for final delivery failed", zap.Error(err), zap.Int64("questionnaire_id", video.QuestionnaireId))
			return nil, err
		}
		if q.UserID <= 0 {
			u.log.Warn("questionnaire has no user id for delivery", zap.Int64("questionnaire_id", video.QuestionnaireId))
		} else {
			videoPath := base.ResolveStoragePath(entity.Path)
			if videoPath == "" {
				u.log.Error("resolve final video path failed", zap.String("path", entity.Path), zap.Int64("questionnaire_id", video.QuestionnaireId))
				return nil, fmt.Errorf("invalid video storage path")
			}

			finalPhotoSet := make(map[string]struct{})
			finalPhotoPaths := make([]string, 0)

			if photos, err := u.photoRepo.GetPhotosQuestionnaire(ctx, video.QuestionnaireId, ""); err != nil {
				u.log.Warn("load questionnaire photos failed", zap.Error(err), zap.Int64("questionnaire_id", video.QuestionnaireId))
			} else {
				for _, photo := range photos {
					if !strings.EqualFold(strings.TrimSpace(photo.TypePhoto), "send") {
						continue
					}
					if abs := base.ResolveStoragePath(photo.Path); abs != "" {
						if _, exists := finalPhotoSet[abs]; exists {
							continue
						}
						finalPhotoSet[abs] = struct{}{}
						finalPhotoPaths = append(finalPhotoPaths, abs)
					}
				}
			}
			//
			//if err := u.tg.SendFinalMedia(ctx, q.UserID, video.QuestionnaireId, finalPhotoPaths, videoPath, finalCaption); err != nil {
			//	u.log.Error("send final media failed", zap.Error(err), zap.Int64("user_id", q.UserID), zap.Int64("questionnaire_id", video.QuestionnaireId))
			//	return nil, err
			//}
		}
	}

	return &protos.Status{Ok: true, Message: "created"}, nil
}
