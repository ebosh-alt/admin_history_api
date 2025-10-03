package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"admin_history/internal/entities"
)

type QuestionnaireRepo struct {
	DB *pgxpool.Pool
}

const (
	questionnaireByIDQuery = `SELECT id, user_id, history, status, payment, created_at, answers
FROM questionnaires WHERE id = $1 order by created_at desc`
	//	photosByQuestionnaireIDQuery = `SELECT id, questionnaire_id, path, scene
	//FROM photos WHERE questionnaire_id = $1 ORDER BY id`
	//	generatePhotosByQuestionnaireIDQuery = `SELECT id, questionnaire_id, path, NULL::text as scene
	//FROM generate_photos WHERE questionnaire_id = $1 ORDER BY id`
	listQuestionnairesQuery = `SELECT id, user_id, answers, history, status, payment, created_at
FROM questionnaires
ORDER BY id
LIMIT $1 OFFSET $2;`
	//	insertPhotoQuery = `INSERT INTO photos (questionnaire_id, path, scene, created_at)
	//VALUES ($1, $2, $3, NOW())
	//ON CONFLICT (questionnaire_id, path) DO UPDATE
	//SET scene = COALESCE(EXCLUDED.scene, photos.scene);`
	//	insertGeneratePhotoQuery = `INSERT INTO generate_photos (questionnaire_id, path, created_at)
	//VALUES ($1, $2, NOW())
	//ON CONFLICT (questionnaire_id, path) DO NOTHING;`
)

func (r *Repository) GetQuestionnaire(ctx context.Context, q *entities.Questionnaire) (*entities.Questionnaire, error) {
	qDTO := q.ToDTO()
	if err := r.DB.QueryRow(ctx, questionnaireByIDQuery, qDTO.ID).Scan(
		&q.ID, &q.UserID, &q.History, &q.Status, &q.Payment, &q.CreatedAt, &q.Answers,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, pgx.ErrNoRows
		}
		return nil, fmt.Errorf("get questionnaire: %w", err)
	}

	return q, nil
}

func (r *Repository) GetQuestionnairesList(
	ctx context.Context,
	page, limit int32,
	f entities.QuestionnaireFilter,
) ([]entities.Questionnaire, error) {
	if page < 1 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}
	if limit > 500 {
		limit = 500
	}
	offset := (page - 1) * limit

	sb := strings.Builder{}
	sb.WriteString(`
SELECT id, user_id, answers, history, status, payment, created_at
FROM questionnaires
WHERE 1=1`)

	args := make([]any, 0, 6)
	i := 0

	if f.Payment != nil {
		i++
		sb.WriteString(fmt.Sprintf(" AND payment = $%d", i))
		args = append(args, *f.Payment)
	}
	if f.Status != nil {
		i++
		sb.WriteString(fmt.Sprintf(" AND status = $%d", i))
		args = append(args, *f.Status)
	}
	if f.DateFrom != nil {
		i++
		sb.WriteString(fmt.Sprintf(" AND created_at >= $%d", i))
		args = append(args, *f.DateFrom)
	}
	if f.DateTo != nil {
		i++
		sb.WriteString(fmt.Sprintf(" AND created_at < $%d", i))
		args = append(args, *f.DateTo)
	}

	i++
	sb.WriteString(fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d", i))
	args = append(args, limit)
	i++
	sb.WriteString(fmt.Sprintf(" OFFSET $%d", i))
	args = append(args, offset)

	rows, err := r.DB.Query(ctx, sb.String(), args...)
	if err != nil {
		return nil, fmt.Errorf("list questionnaires: %w", err)
	}
	defer rows.Close()

	out := make([]entities.Questionnaire, 0, limit)
	for rows.Next() {
		var item entities.Questionnaire
		if err := rows.Scan(
			&item.ID,
			&item.UserID,
			&item.Answers,
			&item.History,
			&item.Status,
			&item.Payment,
			&item.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan questionnaire: %w", err)
		}
		out = append(out, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *Repository) UpdateQuestionnaire(ctx context.Context, q *entities.Questionnaire) error {
	if q == nil || q.ID == 0 {
		return fmt.Errorf("invalid questionnaire")
	}

	set := []string{
		"user_id = $2",
		"history = $3",
		"status  = $4",
		"payment = $5",
	}
	args := []any{
		q.ID,
		q.UserID,
		q.History,
		q.Status,
		q.Payment,
	}
	i := 5

	if q.Answers != nil {
		i++
		set = append(set, fmt.Sprintf("answers = $%d", i))
		args = append(args, q.Answers)
	}

	sql := fmt.Sprintf(`UPDATE questionnaires SET %s WHERE id = $1`, strings.Join(set, ", "))

	tag, err := r.DB.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("update questionnaires: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}
