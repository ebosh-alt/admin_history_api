package questionnaire

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"admin_history/internal/entities"
	"admin_history/internal/repository"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	questionnaireByIDQuery = `SELECT id, user_id, history, storyboard, status, payment, created_at, answers
FROM questionnaires WHERE id = $1 order by created_at desc`
	listQuestionnairesQuery = `SELECT id, user_id, answers, history, storyboard, status, payment, created_at
FROM questionnaires
ORDER BY id
LIMIT $1 OFFSET $2;`
	updateQuestionnaireStatusQuery = `UPDATE questionnaires SET status = $2 WHERE id = $1`
)

type Repo struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) repository.QuestionnaireRepository {
	return &Repo{db: db}
}

func (r *Repo) GetQuestionnaire(ctx context.Context, q *entities.Questionnaire) (*entities.Questionnaire, error) {
	qDTO := q.ToDTO()
	var storyboard sql.NullString
	if err := r.db.QueryRow(ctx, questionnaireByIDQuery, qDTO.ID).Scan(
		&q.ID, &q.UserID, &q.History, &storyboard, &q.Status, &q.Payment, &q.CreatedAt, &q.Answers,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, pgx.ErrNoRows
		}
		return nil, fmt.Errorf("get questionnaire: %w", err)
	}
	if storyboard.Valid {
		q.Storyboard = storyboard.String
	} else {
		q.Storyboard = ""
	}

	return q, nil
}

func (r *Repo) CountQuestionnaires(ctx context.Context, f entities.QuestionnaireFilter) (int64, error) {
	var (
		args  []any
		where []string
		i     = 1
	)

	if f.UserID != nil {
		where = append(where, fmt.Sprintf("user_id = $%d", i))
		args = append(args, *f.UserID)
		i++
	}
	if f.Payment != nil {
		where = append(where, fmt.Sprintf("payment = $%d", i))
		args = append(args, *f.Payment)
		i++
	}
	if f.Status != nil {
		where = append(where, fmt.Sprintf("status = $%d", i))
		args = append(args, *f.Status)
		i++
	}
	if f.DateFrom != nil {
		where = append(where, fmt.Sprintf("created_at >= $%d", i))
		args = append(args, *f.DateFrom)
		i++
	}
	if f.DateTo != nil {
		where = append(where, fmt.Sprintf("created_at < $%d", i))
		args = append(args, *f.DateTo)
		i++
	}

	whereSQL := ""
	if len(where) > 0 {
		whereSQL = "WHERE " + strings.Join(where, " AND ")
	}

	var total int64
	q := "SELECT COUNT(*) FROM questionnaires " + whereSQL
	if err := r.db.QueryRow(ctx, q, args...).Scan(&total); err != nil {
		return 0, err
	}
	return total, nil
}

func (r *Repo) GetQuestionnairesList(
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
SELECT id, user_id, answers, history, storyboard, status, payment, created_at
FROM questionnaires
WHERE 1=1`)

	args := make([]any, 0, 7)
	i := 0

	if f.UserID != nil {
		i++
		sb.WriteString(fmt.Sprintf(" AND user_id = $%d", i))
		args = append(args, *f.UserID)
	}
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
	rows, err := r.db.Query(ctx, sb.String(), args...)
	if err != nil {
		return nil, fmt.Errorf("list questionnaires: %w", err)
	}
	defer rows.Close()
	out := make([]entities.Questionnaire, 0, limit)
	for rows.Next() {
		var item entities.Questionnaire
		var storyboard sql.NullString
		if err := rows.Scan(
			&item.ID,
			&item.UserID,
			&item.Answers,
			&item.History,
			&storyboard,
			&item.Status,
			&item.Payment,
			&item.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan questionnaire: %w", err)
		}
		if storyboard.Valid {
			item.Storyboard = storyboard.String
		}
		out = append(out, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *Repo) UpdateQuestionnaire(ctx context.Context, q *entities.Questionnaire) error {
	if q == nil || q.ID == 0 {
		return fmt.Errorf("invalid questionnaire")
	}

	set := []string{
		"user_id = $2",
		"history = $3",
		"status  = $4",
		"payment = $5",
		"storyboard = $6",
	}
	args := []any{
		q.ID,
		q.UserID,
		q.History,
		q.Status,
		q.Payment,
		q.Storyboard,
	}
	i := len(args)

	if q.Answers != nil {
		i++
		set = append(set, fmt.Sprintf("answers = $%d", i))
		args = append(args, q.Answers)
	}

	sql := fmt.Sprintf(`UPDATE questionnaires SET %s WHERE id = $1`, strings.Join(set, ", "))

	tag, err := r.db.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("update questionnaires: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *Repo) SetQuestionnaireStatus(ctx context.Context, questionnaireID int64, status bool) error {
	if questionnaireID <= 0 {
		return fmt.Errorf("invalid questionnaire id")
	}

	tag, err := r.db.Exec(ctx, updateQuestionnaireStatusQuery, questionnaireID, status)
	if err != nil {
		return fmt.Errorf("set questionnaire status: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

var _ repository.QuestionnaireRepository = (*Repo)(nil)
