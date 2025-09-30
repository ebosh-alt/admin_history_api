package postgres

import (
	"admin_history/internal/entities"
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
)

const (
	userByIDQuery   = `SELECT id, username, status, accepted_offer, created_at FROM users WHERE id = $1;`
	listUserQuery   = `SELECT id, username, status, accepted_offer, created_at FROM users ORDER BY id LIMIT $1 OFFSET $2;`
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

func (r *Repository) UsersList(ctx context.Context, page int32, limit int32) ([]entities.User, error) {
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
	rows, err := r.DB.Query(ctx, listUserQuery,
		limit,
		offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []entities.User
	for rows.Next() {
		var u entities.User
		err := rows.Scan(
			&u.ID, &u.Username, &u.Status, &u.AcceptedOffer, &u.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		users = append(users, u)
	}
	return users, nil

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
