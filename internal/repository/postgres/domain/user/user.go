package user

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
	userByIDQuery   = `SELECT id, username, language, ref_boss_id, status, accepted_offer, created_at, promocode, age, gender, map_binding FROM users WHERE id = $1;`
	updateUserQuery = `UPDATE users SET
  username        = COALESCE($2, username),
  language        = COALESCE($3, language),
  ref_boss_id     = COALESCE($4, ref_boss_id),
  status          = COALESCE($5, status),
  accepted_offer  = COALESCE($6, accepted_offer),
  promocode       = COALESCE($7, promocode),
  age             = COALESCE($8, age),
  gender          = COALESCE($9, gender),
  map_binding     = COALESCE($10, map_binding)
WHERE id = $1
RETURNING id, username, language, ref_boss_id, status, accepted_offer, created_at, promocode, age, gender, map_binding;
`
)

type Repo struct {
	db *pgxpool.Pool
}

func appendPromocodeCondition(sb *strings.Builder, column string, idx *int, args *[]any, promoPtr *string) {
	if promoPtr == nil {
		return
	}

	promo := strings.TrimSpace(*promoPtr)
	if promo == "" {
		return
	}

	options := []string{promo}
	if alt := strings.ReplaceAll(promo, " ", "+"); alt != promo {
		options = append(options, alt)
	}

	sb.WriteString(" AND (")
	for n, opt := range options {
		if n > 0 {
			sb.WriteString(" OR ")
		}
		*idx++
		sb.WriteString(fmt.Sprintf("%s ILIKE $%d", column, *idx))
		*args = append(*args, opt)
	}
	sb.WriteString(")")
}

func New(db *pgxpool.Pool) repository.UserRepository {
	return &Repo{db: db}
}

func (r *Repo) GetUser(ctx context.Context, user *entities.User) (*entities.User, error) {
	userDTO := &entities.UserDTO{}
	err := r.db.QueryRow(ctx, userByIDQuery, user.ID).Scan(
		&userDTO.ID, &userDTO.Username, &userDTO.Language, &userDTO.RefBossID, &userDTO.Status, &userDTO.AcceptedOffer, &userDTO.CreatedAt, &userDTO.Promocode, &userDTO.Age, &userDTO.Gender, &userDTO.MapBinding,
	)
	if err != nil {
		return nil, err
	}
	return userDTO.ToEntity(), nil
}

func (r *Repo) CountUsers(ctx context.Context, f entities.UsersFilter) (int64, error) {
	sb := strings.Builder{}
	sb.WriteString(`SELECT COUNT(*) FROM users WHERE 1=1`)
	args := make([]any, 0, 8)
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
	appendPromocodeCondition(&sb, "promocode", &i, &args, f.Promocode)
	if f.AgeFrom != nil {
		i++
		sb.WriteString(fmt.Sprintf(" AND age >= $%d", i))
		args = append(args, *f.AgeFrom)
	}
	if f.AgeTo != nil {
		i++
		sb.WriteString(fmt.Sprintf(" AND age <= $%d", i))
		args = append(args, *f.AgeTo)
	}
	if f.Gender != nil {
		i++
		sb.WriteString(fmt.Sprintf(" AND gender = $%d", i))
		args = append(args, *f.Gender)
	}
	if f.MapBinding != nil {
		i++
		sb.WriteString(fmt.Sprintf(" AND map_binding = $%d", i))
		args = append(args, *f.MapBinding)
	}

	var total int64
	if err := r.db.QueryRow(ctx, sb.String(), args...).Scan(&total); err != nil {
		return 0, fmt.Errorf("count users: %w", err)
	}
	return total, nil
}

func (r *Repo) UsersList(
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
  u.id, u.username, u.language, u.ref_boss_id, u.status, u.accepted_offer, u.created_at, u.promocode, u.age, u.gender, u.map_binding,
  COALESCE(COUNT(q.id), 0)                                        AS q_total,
  COALESCE(COUNT(*) FILTER (WHERE q.payment = TRUE), 0)           AS q_paid,
  COALESCE(COUNT(*) FILTER (WHERE q.payment = FALSE), 0)          AS q_unpaid
FROM users u
LEFT JOIN questionnaires q ON q.user_id = u.id
WHERE 1=1`)

	args := make([]any, 0, 12)
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
	appendPromocodeCondition(&sb, "u.promocode", &i, &args, f.Promocode)
	if f.AgeFrom != nil {
		i++
		sb.WriteString(fmt.Sprintf(" AND u.age >= $%d", i))
		args = append(args, *f.AgeFrom)
	}
	if f.AgeTo != nil {
		i++
		sb.WriteString(fmt.Sprintf(" AND u.age <= $%d", i))
		args = append(args, *f.AgeTo)
	}
	if f.Gender != nil {
		i++
		sb.WriteString(fmt.Sprintf(" AND u.gender = $%d", i))
		args = append(args, *f.Gender)
	}
	if f.MapBinding != nil {
		i++
		sb.WriteString(fmt.Sprintf(" AND u.map_binding = $%d", i))
		args = append(args, *f.MapBinding)
	}

	sb.WriteString(` GROUP BY u.id, u.username, u.language, u.ref_boss_id, u.status, u.accepted_offer, u.created_at, u.promocode, u.age, u.gender, u.map_binding`)
	i++
	sb.WriteString(fmt.Sprintf(` ORDER BY u.id LIMIT $%d`, i))
	args = append(args, limit)
	i++
	sb.WriteString(fmt.Sprintf(` OFFSET $%d`, i))
	args = append(args, offset)

	rows, err := r.db.Query(ctx, sb.String(), args...)
	if err != nil {
		return nil, fmt.Errorf("users list with stats: %w", err)
	}
	defer rows.Close()

	out := make([]entities.User, 0, limit)
	for rows.Next() {
		var u entities.User
		if err := rows.Scan(
			&u.ID, &u.Username, &u.Language, &u.RefBossID, &u.Status, &u.AcceptedOffer, &u.CreatedAt, &u.Promocode, &u.Age, &u.Gender, &u.MapBinding,
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

func (r *Repo) UpdateUser(ctx context.Context, user *entities.User) error {
	userDTO := user.ToDTO()
	tag, err := r.db.Exec(ctx, updateUserQuery,
		userDTO.ID,
		userDTO.Username,
		userDTO.Language,
		userDTO.RefBossID,
		userDTO.Status,
		userDTO.AcceptedOffer,
		userDTO.Promocode,
		userDTO.Age,
		userDTO.Gender,
		userDTO.MapBinding,
	)
	if err != nil {
		return fmt.Errorf("update users: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

var _ repository.UserRepository = (*Repo)(nil)
