package postgres

import (
	"admin_history/internal/entities"
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"strings"
)

const (
	userByIDQuery   = `SELECT id, username, status, accepted_offer, created_at FROM users WHERE id = $1;`
	updateUserQuery = `UPDATE users SET
  username        = COALESCE($2, username),
  status          = COALESCE($3, status),
  accepted_offer  = COALESCE($4, accepted_offer)
WHERE id = $1
RETURNING id, username, status, accepted_offer, created_at;
`
)

func (r *Repository) GetUser(ctx context.Context, user *entities.User) (*entities.User, error) {
	userDTO := &entities.UserDTO{}
	err := r.DB.QueryRow(ctx, userByIDQuery, user.ID).Scan(
		&userDTO.ID, &userDTO.Username, &userDTO.Status, &userDTO.AcceptedOffer, &userDTO.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return userDTO.ToEntity(), nil
}

func (r *Repository) CountUsers(ctx context.Context, f entities.UsersFilter) (int64, error) {
	sb := strings.Builder{}
	sb.WriteString(`SELECT COUNT(*) FROM users WHERE 1=1`)
	args := make([]any, 0, 4)
	i := 0

	if f.Status != nil {
		i++
		sb.WriteString(fmt.Sprintf(" AND status = $%d", i))
		args = append(args, *f.Status)
	}
	if f.AcceptedOffer != nil {
		i++
		sb.WriteString(fmt.Sprintf(" AND accepted_offer = $%d", i))
		args = append(args, *f.AcceptedOffer)
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
	if err := r.DB.QueryRow(ctx, sb.String(), args...).Scan(&total); err != nil {
		return 0, fmt.Errorf("count users: %w", err)
	}
	return total, nil
}

func (r *Repository) UsersList(
	ctx context.Context,
	page, limit int32,
	f entities.UsersFilter,
) ([]entities.User, error) {
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
SELECT
  u.id, u.username, u.status, u.accepted_offer, u.created_at,
  COALESCE(COUNT(q.id), 0)                                        AS q_total,
  COALESCE(COUNT(*) FILTER (WHERE q.payment = TRUE), 0)           AS q_paid,
  COALESCE(COUNT(*) FILTER (WHERE q.payment = FALSE), 0)          AS q_unpaid
FROM users u
LEFT JOIN questionnaires q ON q.user_id = u.id
WHERE 1=1`)

	args := make([]any, 0, 8)
	i := 0

	if f.Status != nil {
		i++
		sb.WriteString(fmt.Sprintf(" AND u.status = $%d", i))
		args = append(args, *f.Status)
	}
	if f.AcceptedOffer != nil {
		i++
		sb.WriteString(fmt.Sprintf(" AND u.accepted_offer = $%d", i))
		args = append(args, *f.AcceptedOffer)
	}
	if f.DateFrom != nil {
		i++
		sb.WriteString(fmt.Sprintf(" AND u.created_at >= $%d", i))
		args = append(args, *f.DateFrom)
	}
	if f.DateTo != nil {
		i++
		sb.WriteString(fmt.Sprintf(" AND u.created_at < $%d", i))
		args = append(args, *f.DateTo)
	}

	sb.WriteString(` GROUP BY u.id, u.username, u.status, u.accepted_offer, u.created_at`)
	i++
	sb.WriteString(fmt.Sprintf(` ORDER BY u.id LIMIT $%d`, i))
	args = append(args, limit)
	i++
	sb.WriteString(fmt.Sprintf(` OFFSET $%d`, i))
	args = append(args, offset)

	rows, err := r.DB.Query(ctx, sb.String(), args...)
	if err != nil {
		return nil, fmt.Errorf("users list with stats: %w", err)
	}
	defer rows.Close()

	out := make([]entities.User, 0, limit)
	for rows.Next() {
		var u entities.User
		if err := rows.Scan(
			&u.ID, &u.Username, &u.Status, &u.AcceptedOffer, &u.CreatedAt,
			&u.QTotal, &u.QPaid, &u.QUnpaid,
		); err != nil {
			return nil, fmt.Errorf("scan users with stats: %w", err)
		}
		out = append(out, u)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *Repository) UpdateUser(ctx context.Context, user *entities.User) error {
	userDTO := user.ToDTO()
	tag, err := r.DB.Exec(ctx, updateUserQuery,
		userDTO.ID,
		userDTO.Username,
		userDTO.Status,
		userDTO.AcceptedOffer,
	)
	if err != nil {
		return fmt.Errorf("update users: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

var _ InterfaceRepo = (*Repository)(nil)
