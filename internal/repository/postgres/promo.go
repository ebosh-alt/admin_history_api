package postgres

import (
	"admin_history/internal/entities"
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
)

const (
	promoCodeByIDQuery = `SELECT id, value, number_uses, status, percent, description 
FROM promo_codes WHERE id = $1;`

	createPromoCodeQuery = `INSERT INTO promo_codes (value, number_uses, status, percent, description) 
VALUES ($1, $2, $3, $4, $5) RETURNING id;`

	updatePromoCodeQuery = `UPDATE promo_codes SET
  value = COALESCE($2, value),
  number_uses = COALESCE($3, number_uses),
  status = COALESCE($4, status),
  percent = COALESCE($5, percent),
  description = COALESCE($6, description)
WHERE id = $1
RETURNING id, value, number_uses, status, percent, description;`
)

func (r *Repository) GetPromoCode(ctx context.Context, promoCode *entities.PromoCode) (*entities.PromoCode, error) {
	promoCodeDTO := promoCode.ToDTO()
	err := r.DB.QueryRow(ctx, promoCodeByIDQuery, promoCodeDTO.ID).Scan(
		&promoCode.ID, &promoCode.Value, &promoCode.NumberUses, &promoCode.Status, &promoCode.Percent, &promoCode.Description,
	)
	if err != nil {
		if pgx.ErrNoRows == err {
			return nil, pgx.ErrNoRows
		}
		return nil, fmt.Errorf("get promo code: %w", err)
	}
	return promoCode, nil
}

func (r *Repository) CountPromoCodes(ctx context.Context, f entities.PromoCodeFilter) (int64, error) {
	sb := strings.Builder{}
	sb.WriteString(`SELECT COUNT(*) FROM promo_codes WHERE 1=1`)
	args := make([]any, 0, 1)
	i := 0

	if f.Status != nil {
		i++
		sb.WriteString(fmt.Sprintf(" AND status = $%d", i))
		args = append(args, *f.Status)
	}

	var total int64
	if err := r.DB.QueryRow(ctx, sb.String(), args...).Scan(&total); err != nil {
		return 0, fmt.Errorf("count promo codes: %w", err)
	}
	return total, nil
}

func (r *Repository) PromoCodesList(
	ctx context.Context,
	page, limit int32,
	f entities.PromoCodeFilter,
) ([]entities.PromoCode, error) {
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
SELECT id, value, number_uses, status, percent, description
FROM promo_codes
WHERE 1=1`)

	args := make([]any, 0, 4)
	i := 0

	if f.Status != nil {
		i++
		sb.WriteString(fmt.Sprintf(" AND status = $%d", i))
		args = append(args, *f.Status)
	}

	sb.WriteString(` ORDER BY id`)
	i++
	sb.WriteString(fmt.Sprintf(` LIMIT $%d`, i))
	args = append(args, limit)
	i++
	sb.WriteString(fmt.Sprintf(` OFFSET $%d`, i))
	args = append(args, offset)

	rows, err := r.DB.Query(ctx, sb.String(), args...)
	if err != nil {
		return nil, fmt.Errorf("list promo codes: %w", err)
	}
	defer rows.Close()

	out := make([]entities.PromoCode, 0, limit)
	for rows.Next() {
		var promoCode entities.PromoCode
		if err := rows.Scan(
			&promoCode.ID, &promoCode.Value, &promoCode.NumberUses, &promoCode.Status, &promoCode.Percent, &promoCode.Description,
		); err != nil {
			return nil, fmt.Errorf("scan promo code: %w", err)
		}
		out = append(out, promoCode)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *Repository) CreatePromoCode(ctx context.Context, promoCode *entities.PromoCode) error {
	promoCodeDTO := promoCode.ToDTO()

	var id int64
	err := r.DB.QueryRow(ctx, createPromoCodeQuery,
		valOrNil(promoCodeDTO.Value),
		valOrNil(promoCodeDTO.NumberUses),
		valOrNil(promoCodeDTO.Status),
		valOrNil(promoCodeDTO.Percent),
		valOrNil(promoCodeDTO.Description),
	).Scan(&id)
	if err != nil {
		return fmt.Errorf("create promo code: %w", err)
	}

	promoCode.ID = id
	return nil
}

func (r *Repository) UpdatePromoCode(ctx context.Context, promoCode *entities.PromoCode) error {
	if promoCode == nil || promoCode.ID == 0 {
		return fmt.Errorf("invalid promo code")
	}

	dto := promoCode.ToDTO()

	err := r.DB.QueryRow(ctx, updatePromoCodeQuery,
		promoCode.ID,
		valOrNil(dto.Value),
		valOrNil(dto.NumberUses),
		valOrNil(dto.Status),
		valOrNil(dto.Percent),
		valOrNil(dto.Description),
	).Scan(
		&promoCode.ID,
		&promoCode.Value,
		&promoCode.NumberUses,
		&promoCode.Status,
		&promoCode.Percent,
		&promoCode.Description,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return pgx.ErrNoRows
		}
		return fmt.Errorf("update promo code: %w", err)
	}

	return nil
}
