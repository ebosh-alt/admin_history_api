package review

import (
	"admin_history/internal/entities"
	"admin_history/internal/repository"
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	reviewByIDQuery = `SELECT id, user_id, description, created_at 
FROM reviews WHERE id = $1;`
)

type Repo struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) repository.ReviewRepository {
	return &Repo{db: db}
}

func (r *Repo) GetReview(ctx context.Context, review *entities.Review) (*entities.Review, error) {
	reviewDTO := review.ToDTO()
	err := r.db.QueryRow(ctx, reviewByIDQuery, reviewDTO.ID).Scan(
		&review.ID, &review.UserID, &review.Description, &review.CreatedAt,
	)
	if err != nil {
		if pgx.ErrNoRows == err {
			return nil, pgx.ErrNoRows
		}
		return nil, fmt.Errorf("get review: %w", err)
	}
	return review, nil
}

func (r *Repo) CountReviews(ctx context.Context, f entities.ReviewFilter) (int64, error) {
	sb := strings.Builder{}
	sb.WriteString(`SELECT COUNT(*) FROM reviews WHERE 1=1`)
	args := make([]any, 0, 3)
	i := 0

	if f.UserID != nil {
		i++
		sb.WriteString(fmt.Sprintf(" AND user_id = $%d", i))
		args = append(args, *f.UserID)
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

	var total int64
	if err := r.db.QueryRow(ctx, sb.String(), args...).Scan(&total); err != nil {
		return 0, fmt.Errorf("count reviews: %w", err)
	}
	return total, nil
}

func (r *Repo) ReviewsList(
	ctx context.Context,
	page, limit int32,
	f entities.ReviewFilter,
) ([]entities.Review, error) {
	if page < 1 {
		page = 1
	}
	if limit <= 0 {
		limit = 50
	}
	if limit > 500 {
		limit = 500
	}
	offset := (page - 1) * limit

	sb := strings.Builder{}
	sb.WriteString(`
SELECT id, user_id, description, created_at
FROM reviews
WHERE 1=1`)

	args := make([]any, 0, 5)
	i := 0

	if f.UserID != nil {
		i++
		sb.WriteString(fmt.Sprintf(" AND user_id = $%d", i))
		args = append(args, *f.UserID)
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

	sb.WriteString(` ORDER BY created_at DESC`)
	i++
	sb.WriteString(fmt.Sprintf(` LIMIT $%d`, i))
	args = append(args, limit)
	i++
	sb.WriteString(fmt.Sprintf(` OFFSET $%d`, i))
	args = append(args, offset)

	rows, err := r.db.Query(ctx, sb.String(), args...)
	if err != nil {
		return nil, fmt.Errorf("list reviews: %w", err)
	}
	defer rows.Close()

	out := make([]entities.Review, 0, limit)
	for rows.Next() {
		var review entities.Review
		if err := rows.Scan(
			&review.ID, &review.UserID, &review.Description, &review.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan review: %w", err)
		}
		out = append(out, review)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

var _ repository.ReviewRepository = (*Repo)(nil)
