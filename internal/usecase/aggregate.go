package errors

import (
	"go.uber.org/fx"
)

type aggregateDeps struct {
	fx.In

	User          InterfaceUserUsecase
	Questionnaire InterfaceQuestionnaireUsecase
	Photo         InterfacePhotoUsecase
	Video         InterfaceVideoUsecase
	Promo         InterfacePromoCodeUsecase
	Review        InterfaceReviewUsecase
}

type aggregateUsecase struct {
	InterfaceUserUsecase
	InterfaceQuestionnaireUsecase
	InterfacePhotoUsecase
	InterfaceVideoUsecase
	InterfacePromoCodeUsecase
	InterfaceReviewUsecase
}

func newAggregateUsecase(deps aggregateDeps) InterfaceUsecase {
	agg := &aggregateUsecase{
		InterfaceUserUsecase:          deps.User,
		InterfaceQuestionnaireUsecase: deps.Questionnaire,
		InterfacePhotoUsecase:         deps.Photo,
		InterfaceVideoUsecase:         deps.Video,
		InterfacePromoCodeUsecase:     deps.Promo,
		InterfaceReviewUsecase:        deps.Review,
	}

	return agg
}
