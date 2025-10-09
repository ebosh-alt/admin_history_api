package usecase

import (
	"admin_history/internal/entities"
	protos "admin_history/pkg/proto/gen/go"
	"context"
	"errors"
	"fmt"
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
	page, limit := req.GetPage(), req.GetLimit()
	if page < 1 {
		page = 1
	}
	if limit <= 0 {
		limit = 50
	}

	var f entities.UsersFilter
	if req.Status != nil {
		v := req.Status.Value
		f.Status = &v
	}
	if req.AcceptedOffer != nil {
		v := req.AcceptedOffer.Value
		f.AcceptedOffer = &v
	}
	if req.DateFrom != nil {
		t := req.DateFrom.AsTime()
		f.DateFrom = &t
	}
	if req.DateTo != nil {
		t := req.DateTo.AsTime()
		f.DateTo = &t
	}

	items, err := u.Postgres.UsersList(ctx, page, limit, f)
	if err != nil {
		return nil, fmt.Errorf("users list: %w", err)
	}
	countUsers, err := u.Postgres.CountUsers(ctx, f)

	rows := make([]*protos.User, 0, len(items))
	for i := range items {
		it := items[i]
		rows = append(rows, &protos.User{
			Id:            it.ID,
			Username:      it.Username,
			Status:        it.Status,
			AcceptedOffer: it.AcceptedOffer,
			CreatedAt:     timestamppb.New(it.CreatedAt),
			Total:         it.QTotal,
			Paid:          it.QPaid,
			Unpaid:        it.QUnpaid,
		})
	}

	return &protos.UsersListResponse{
		Users: rows,
		Total: countUsers,
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
