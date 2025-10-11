package usecase

import (
	"admin_history/internal/entities"
	protos "admin_history/pkg/proto/gen/go"
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"go.uber.org/zap"
	"io"
)

func (u *Usecase) GetVideosQuestionnaire(ctx context.Context, req *protos.VideoRequest) (*protos.VideoResponse, error) {
	qID := req.GetQuestionnaireId()
	typeVideo := strings.TrimSpace(req.GetType())

	videos, err := u.Postgres.GetVideosQuestionnaire(ctx, qID, typeVideo)
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

func (u *Usecase) UploadVideo(ctx context.Context, r io.Reader, filename string, video *protos.Video) (*protos.Status, error) {
	if video.GetQuestionnaireId() <= 0 {
		return nil, fmt.Errorf("invalid questionnaire_id")
	}

	ext := normalizeVideoExt(filename)
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

	if err := u.Postgres.UploadVideo(ctx, &entity); err != nil {
		if dd := u.st.Remove(ctx, relPath); dd != nil {
			u.log.Warn("cleanup remove file failed", zap.String("path", relPath), zap.Error(dd))
		}
		return nil, err
	}

	return &protos.Status{Ok: true, Message: "created"}, nil
}

func normalizeVideoExt(name string) string {
	ext := strings.ToLower(strings.TrimSpace(filepath.Ext(name)))
	switch ext {
	case ".mp4", ".mov", ".avi", ".mkv", ".webm":
		return ext
	default:
		return ".bin"
	}
}

func normalizeVideoType(t string) string {
	t = strings.ToLower(strings.TrimSpace(t))
	if isAllowedVideoType(t) {
		return t
	}
	return "send"
}

var _ InterfaceVideoUsecase = (*Usecase)(nil)
