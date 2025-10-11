package usecase

import (
	"admin_history/internal/entities"
	protos "admin_history/pkg/proto/gen/go"
	"context"
	"fmt"
	"go.uber.org/zap"
	"io"
	"path/filepath"
	"strings"
)

func (u *Usecase) GetPhotosQuestionnaire(ctx context.Context, req *protos.PhotoRequest) (*protos.PhotoResponse, error) {
	qID := req.QuestionnaireId
	qType := req.Type
	if qType == "" {
		qType = "original"
	}
	photo, err := u.Postgres.GetPhotosQuestionnaire(ctx, qID, qType)
	if err != nil {
		return nil, err
	}
	photoProto := make([]*protos.Photo, 0, len(photo))
	for _, p := range photo {
		pProtos := &protos.Photo{
			QuestionnaireId: p.QuestionnaireID,
			Path:            p.Path,
			Scene:           p.Scene,
			TypePhoto:       p.TypePhoto,
		}
		photoProto = append(photoProto, pProtos)
	}
	return &protos.PhotoResponse{
		Photo: photoProto,
	}, nil
}

func (u *Usecase) UploadPhoto(ctx context.Context, r io.Reader, filename string, photo *protos.Photo) (*protos.Status, error) {
	if photo.QuestionnaireId <= 0 {
		return nil, fmt.Errorf("invalid questionnaire_id")
	}

	ext := normalizeExt(filename)
	relPath, err := u.st.Save(ctx, r, ext)
	if relPath == "" {
		return nil, fmt.Errorf("invalid photo payload")
	}
	if err != nil {
		u.log.Error("storage save failed", zap.Error(err))
		return nil, err
	}
	photo.Path = relPath

	ent := entities.Photo{
		Path:            photo.Path,
		QuestionnaireID: photo.QuestionnaireId,
		Scene:           photo.Scene,
		TypePhoto:       normalizePhotoType(photo.TypePhoto),
	}

	if err := u.Postgres.UploadPhoto(ctx, &ent); err != nil {
		if dd := u.st.Remove(ctx, relPath); dd != nil {
			u.log.Warn("cleanup remove file failed", zap.String("path", relPath), zap.Error(dd))
		}
		return nil, err
	}
	return &protos.Status{Ok: true, Message: "created"}, nil
}

func normalizeExt(name string) string {
	ext := strings.ToLower(strings.TrimSpace(filepath.Ext(name)))
	switch ext {
	case ".jpg", ".jpeg":
		return ".jpg"
	case ".png", ".webp":
		return ext
	default:
		return ".bin"
	}
}

func normalizePhotoType(t string) string {
	t = strings.ToLower(strings.TrimSpace(t))
	if isAllowedPhotoType(t) {
		return t
	}
	return "original"
}

var _ InterfacePhotoUsecase = (*Usecase)(nil)
