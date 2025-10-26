package promo

import (
	"admin_history/internal/repository"
	"admin_history/internal/storage"
	"context"
	"errors"
	"fmt"
	"go.uber.org/zap"

	"admin_history/internal/entities"
	protos "admin_history/pkg/proto/gen/go"

	"github.com/jackc/pgx/v5"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type Usecase struct {
	log  *zap.Logger
	repo repository.PromoCodeRepository
	st   *storage.FS
}

func New(log *zap.Logger, repo repository.PromoCodeRepository, st *storage.FS) *Usecase {
	return &Usecase{
		log:  log,
		repo: repo,
		st:   st,
	}
}
func (u *Usecase) GetPromoCode(ctx context.Context, req *protos.PromoCodeRequest) (*protos.PromoCodeResponse, error) {
	if req == nil || req.Id <= 0 {
		return nil, fmt.Errorf("invalid promo code id")
	}

	promoCode := &entities.PromoCode{ID: req.Id}
	promoCode, err := u.repo.GetPromoCode(ctx, promoCode)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repository.ErrPromoCodeNotFound
		}
		return nil, err
	}

	return &protos.PromoCodeResponse{
		PromoCode: &protos.PromoCode{
			Id:          promoCode.ID,
			Value:       promoCode.Value,
			NumberUses:  wrapperspb.Int32(int32(*promoCode.NumberUses)),
			Status:      wrapperspb.Bool(*promoCode.Status),
			Percent:     promoCode.Percent,
			Description: promoCode.Description,
		},
	}, nil
}

func (u *Usecase) PromoCodesList(ctx context.Context, req *protos.PromoCodesListRequest) (*protos.PromoCodesListResponse, error) {
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

	var f entities.PromoCodeFilter
	if req.Status != nil {
		v := req.Status.Value
		f.Status = &v
	}

	items, err := u.repo.PromoCodesList(ctx, page, limit, f)
	if err != nil {
		return nil, fmt.Errorf("promo codes list: %w", err)
	}

	count, err := u.repo.CountPromoCodes(ctx, f)
	if err != nil {
		return nil, fmt.Errorf("count promo codes: %w", err)
	}

	promoCodes := make([]*protos.PromoCode, 0, len(items))
	for i := range items {
		item := items[i]
		promoCode := &protos.PromoCode{
			Id:          item.ID,
			Value:       item.Value,
			Percent:     item.Percent,
			Description: item.Description,
		}

		if item.NumberUses != nil {
			promoCode.NumberUses = wrapperspb.Int32(int32(*item.NumberUses))
		}
		if item.Status != nil {
			promoCode.Status = wrapperspb.Bool(*item.Status)
		}

		promoCodes = append(promoCodes, promoCode)
	}

	return &protos.PromoCodesListResponse{
		PromoCodes: promoCodes,
		Total:      count,
	}, nil
}

func (u *Usecase) CreatePromoCode(ctx context.Context, req *protos.CreatePromoCodeRequest) (*protos.Status, error) {
	if req == nil || req.PromoCode == nil {
		return nil, fmt.Errorf("invalid request")
	}

	promoCode := &entities.PromoCode{
		Value:       req.PromoCode.Value,
		Percent:     req.PromoCode.Percent,
		Description: req.PromoCode.Description,
	}

	if req.PromoCode.NumberUses != nil {
		numberUses := int(req.PromoCode.NumberUses.Value)
		promoCode.NumberUses = &numberUses
	}

	if req.PromoCode.Status != nil {
		status := req.PromoCode.Status.Value
		promoCode.Status = &status
	}

	err := u.repo.CreatePromoCode(ctx, promoCode)
	if err != nil {
		return nil, fmt.Errorf("create promo code: %w", err)
	}

	return &protos.Status{
		Ok:      true,
		Message: "promo code created successfully",
	}, nil
}

func (u *Usecase) UpdatePromoCode(ctx context.Context, req *protos.UpdatePromoCodeRequest) (*protos.Status, error) {
	if req == nil || req.PromoCode == nil || req.PromoCode.Id <= 0 {
		return nil, fmt.Errorf("invalid request")
	}

	if req.PromoCode.Value == "" {
		return nil, fmt.Errorf("promo code value is required")
	}

	if req.PromoCode.Percent <= 0 {
		return nil, fmt.Errorf("promo code percent must be positive")
	}

	if req.PromoCode.Description == "" {
		return nil, fmt.Errorf("promo code description is required")
	}

	promoCode := &entities.PromoCode{
		ID:          req.PromoCode.Id,
		Value:       req.PromoCode.Value,
		Percent:     req.PromoCode.Percent,
		Description: req.PromoCode.Description,
	}

	if req.PromoCode.NumberUses != nil {
		numberUses := int(req.PromoCode.NumberUses.Value)
		promoCode.NumberUses = &numberUses
	}

	if req.PromoCode.Status != nil {
		status := req.PromoCode.Status.Value
		promoCode.Status = &status
	}

	err := u.repo.UpdatePromoCode(ctx, promoCode)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repository.ErrPromoCodeNotFound
		}
		return nil, fmt.Errorf("update promo code: %w", err)
	}

	return &protos.Status{
		Ok:      true,
		Message: "promo code updated successfully",
	}, nil
}
