package errors

import (
	"admin_history/internal/usecase/domain/photo"
	"admin_history/internal/usecase/domain/promo"
	"admin_history/internal/usecase/domain/questionnaire"
	"admin_history/internal/usecase/domain/review"
	"admin_history/internal/usecase/domain/user"
	"admin_history/internal/usecase/domain/video"
	"go.uber.org/fx"
)

func Module() fx.Option {
	return fx.Module(
		"usecase",
		fx.Provide(
			fx.Annotate(photo.New, fx.As(new(InterfacePhotoUsecase))),
			fx.Annotate(promo.New, fx.As(new(InterfacePromoCodeUsecase))),
			fx.Annotate(questionnaire.New, fx.As(new(InterfaceQuestionnaireUsecase))),
			fx.Annotate(review.New, fx.As(new(InterfaceReviewUsecase))),
			fx.Annotate(user.New, fx.As(new(InterfaceUserUsecase))),
			fx.Annotate(video.New, fx.As(new(InterfaceVideoUsecase))),
			fx.Annotate(newAggregateUsecase, fx.As(new(InterfaceUsecase))),
		),
	)
}
