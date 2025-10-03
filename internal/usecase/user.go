package usecase

import (
	"admin_history/internal/entities"
	protos "admin_history/pkg/proto/gen/go"
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (u *Usecase) GetUser(ctx context.Context, req *protos.UserRequest) (*protos.UserResponse, error) {
	userId := req.Id
	if userId == 0 {
		return nil, errors.New("user id is required")
	}
	user := &entities.User{ID: userId}
	user, err := u.Postgres.GetUser(ctx, user)
	if err != nil {
		return nil, errors.New("user not found")
	}
	return &protos.UserResponse{
		User: &protos.User{
			Id:            user.ID,
			Username:      user.Username,
			Status:        user.Status,
			AcceptedOffer: user.AcceptedOffer,
			CreatedAt:     timestamppb.New(user.CreatedAt),
		},
	}, err

}

func (u *Usecase) UsersList(ctx context.Context, req *protos.UsersListRequest) (*protos.UsersListResponse, error) {
	page := req.Page
	limit := req.Limit
	if page == 0 {
		page = 1
	}
	if limit == 0 {
		limit = 50
	}

	usersList, err := u.Postgres.UsersList(ctx, page, limit)
	if err != nil {
		return nil, errors.New("users not found")
	}
	protosUsers := make([]*protos.User, len(usersList))
	for i := 0; i < len(usersList); i++ {
		user := usersList[i]
		protosUsers[i] = &protos.User{
			Id:            user.ID,
			Username:      user.Username,
			Status:        user.Status,
			AcceptedOffer: user.AcceptedOffer,
			CreatedAt:     timestamppb.New(user.CreatedAt),
		}
	}
	return &protos.UsersListResponse{
		Users: protosUsers,
	}, nil
}

func (u *Usecase) UpdateUser(ctx context.Context, req *protos.UpdateUserRequest) (*protos.Status, error) {
	user := &entities.User{
		ID:            req.User.Id,
		Username:      req.User.Username,
		Status:        req.User.Status,
		AcceptedOffer: req.User.AcceptedOffer,
	}
	err := u.Postgres.UpdateUser(ctx, user)
	if errors.Is(err, pgx.ErrNoRows) {
		return &protos.Status{
			Ok:      true,
			Message: "has no rows affected",
		}, nil
	} else if err != nil {
		return nil, errors.New("user not updated")
	}

	return &protos.Status{
		Ok:      true,
		Message: "",
	}, nil
}
