package postgres

import (
	"admin_history/internal/entities"
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

const (
	listVideosByQIDQuery = `
		SELECT questionnaire_id, path, type_video, created_at
		FROM videos
		WHERE questionnaire_id = $1
		ORDER BY created_at NULLS LAST, path
	`

	listVideosByQIDTypeQuery = `
		SELECT questionnaire_id, path, type_video, created_at
		FROM videos
		WHERE questionnaire_id = $1 AND type_video = $2
		ORDER BY created_at NULLS LAST, path
	`

	insertVideoQuery = `
		INSERT INTO videos (questionnaire_id, path, type_video, created_at)
		VALUES ($1, $2, $3, NOW())
		RETURNING id, created_at
	`
)

func (r *Repository) GetVideosQuestionnaire(ctx context.Context, questionnaireID int64, typeVideo string) ([]entities.Video, error) {
	var (
		rows pgx.Rows
		err  error
	)

	switch {
	case typeVideo == "", typeVideo == "all":
		rows, err = r.DB.Query(ctx, listVideosByQIDQuery, questionnaireID)
	default:
		rows, err = r.DB.Query(ctx, listVideosByQIDTypeQuery, questionnaireID, typeVideo)
	}
	if err != nil {
		return nil, fmt.Errorf("get videos: %w", err)
	}
	defer rows.Close()

	items := make([]entities.Video, 0, 16)
	for rows.Next() {
		var v entities.Video
		if scanErr := rows.Scan(&v.QuestionnaireID, &v.Path, &v.TypeVideo, &v.CreatedAt); scanErr != nil {
			return nil, fmt.Errorf("scan video: %w", scanErr)
		}
		items = append(items, v)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *Repository) UploadVideo(ctx context.Context, video *entities.Video) error {
	if video == nil {
		return fmt.Errorf("video payload is nil")
	}

	vDTO := video.ToDTO()
	return r.DB.QueryRow(ctx, insertVideoQuery,
		vDTO.QuestionnaireID,
		vDTO.Path,
		vDTO.TypeVideo,
	).Scan(&vDTO.ID, &vDTO.CreatedAt)
}

var _ InterfaceVideoRepo = (*Repository)(nil)
