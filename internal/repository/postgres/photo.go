package postgres

import (
	"admin_history/internal/entities"
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
)

const (
	listPhotosByQIDQuery = `
		SELECT questionnaire_id, path, scene, type, created_at
		FROM photos
		WHERE questionnaire_id = $1
		ORDER BY created_at NULLS LAST, path
	`

	listPhotosByQIDTypeQuery = `
		SELECT questionnaire_id, path, scene, type, created_at
		FROM photos
		WHERE questionnaire_id = $1 AND type = $2
		ORDER BY created_at NULLS LAST, path
	`

	insertPhotoQuery = `INSERT INTO photos (questionnaire_id, path, scene, type, created_at) VALUES ($1, $2, $3, $4, NOW()) returning id`
)

func (r *Repository) GetPhotosQuestionnaire(ctx context.Context, questionnaireID int64, typePhoto string) ([]entities.Photo, error) {

	var (
		rows pgx.Rows
		err  error
	)

	if typePhoto == "" || typePhoto == "all" {
		const q = listPhotosByQIDQuery
		rows, err = r.DB.Query(ctx, q, questionnaireID)
	} else {
		const q = listPhotosByQIDTypeQuery
		rows, err = r.DB.Query(ctx, q, questionnaireID, typePhoto)
	}
	if err != nil {
		return nil, fmt.Errorf("get photos: %w", err)
	}
	defer rows.Close()

	out := make([]entities.Photo, 0, 64)
	for rows.Next() {
		var p entities.Photo
		if err := rows.Scan(&p.QuestionnaireID, &p.Path, &p.Scene, &p.TypePhoto); err != nil {
			return nil, fmt.Errorf("scan photo: %w", err)
		}
		out = append(out, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *Repository) UploadPhoto(ctx context.Context, photo *entities.Photo) error {
	pDTO := photo.ToDTO()
	err := r.DB.QueryRow(ctx, insertPhotoQuery,
		pDTO.QuestionnaireID,
		pDTO.Path,
		pDTO.Scene,
		pDTO.TypePhoto).Scan(&pDTO.ID)
	if err != nil {
		return err
	}

	return nil
}

var _ InterfacePhotoRepo = (*Repository)(nil)
