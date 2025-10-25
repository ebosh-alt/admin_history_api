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

func (u *Usecase) GetReview(ctx context.Context, req *protos.ReviewRequest) (*protos.ReviewResponse, error) {
	if req == nil || req.Id <= 0 {
		return nil, fmt.Errorf("invalid review id")
	}

	review := &entities.Review{ID: req.Id}
	review, err := u.reviews.GetReview(ctx, review)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("review not found")
		}
		return nil, err
	}

	return &protos.ReviewResponse{
		Review: &protos.Review{
			Id:          review.ID,
			UserId:      review.UserID,
			Description: review.Description,
			CreatedAt:   timestamppb.New(review.CreatedAt),
		},
	}, nil
}

func (u *Usecase) ReviewsList(ctx context.Context, req *protos.ReviewsListRequest) (*protos.ReviewsListResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("nil request")
	}

	page, limit := req.GetPage(), req.GetLimit()
	if page < 1 {
		page = 1
	}
	if limit <= 0 {
		limit = 50
	}

	var f entities.ReviewFilter
	if req.UserId != nil {
		f.UserID = &req.UserId.Value
	}
	if req.DateFrom != nil {
		t := req.DateFrom.AsTime()
		f.DateFrom = &t
	}
	if req.DateTo != nil {
		t := req.DateTo.AsTime()
		f.DateTo = &t
	}

	items, err := u.reviews.ReviewsList(ctx, page, limit, f)
	if err != nil {
		return nil, fmt.Errorf("reviews list: %w", err)
	}

	count, err := u.reviews.CountReviews(ctx, f)
	if err != nil {
		return nil, fmt.Errorf("count reviews: %w", err)
	}

	reviews := make([]*protos.Review, 0, len(items))
	for i := range items {
		item := items[i]
		reviews = append(reviews, &protos.Review{
			Id:          item.ID,
			UserId:      item.UserID,
			Description: item.Description,
			CreatedAt:   timestamppb.New(item.CreatedAt),
		})
	}

	return &protos.ReviewsListResponse{
		Reviews: reviews,
		Total:   count,
	}, nil
}
