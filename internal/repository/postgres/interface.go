package postgres

import (
	"admin_history/internal/entities"
	protos "admin_history/pkg/proto/gen/go"
	"context"
)

type InterfaceRepo interface {
	InterfaceUserRepo
	InterfaceQuestionnaireRepo
	InterfacePhotoRepo
	InterfaceVideoRepo
	InterfacePromoCodeRepo
	InterfaceReviewRepo
}

type InterfaceUserRepo interface {
	GetUser(ctx context.Context, user *entities.User) (*entities.User, error)
	UsersList(ctx context.Context, page int32, limit int32, f entities.UsersFilter) ([]entities.User, error)
	UpdateUser(ctx context.Context, user *entities.User) error
	CountUsers(ctx context.Context, f entities.UsersFilter) (int64, error)
}

type InterfaceQuestionnaireRepo interface {
	GetQuestionnaire(ctx context.Context, questionnaire *entities.Questionnaire) (*entities.Questionnaire, error)
	GetQuestionnairesList(ctx context.Context, page int32, limit int32, f entities.QuestionnaireFilter) ([]entities.Questionnaire, error)
	UpdateQuestionnaire(ctx context.Context, req *entities.Questionnaire) error
	CountQuestionnaires(ctx context.Context, f entities.QuestionnaireFilter) (int64, error)
}

type InterfacePhotoRepo interface {
	GetPhotosQuestionnaire(ctx context.Context, questionnaireID int64, typePhoto string) ([]entities.Photo, error)
	UploadPhoto(ctx context.Context, photo *entities.Photo) error
}

type InterfaceVideoRepo interface {
	GetVideosQuestionnaire(ctx context.Context, questionnaireID int64, typeVideo string) ([]entities.Video, error)
	UploadVideo(ctx context.Context, video *entities.Video) error
}

type InterfaceChatRepo interface {
	GetChat(ctx context.Context, req *protos.ChatRequest) (*protos.ChatResponse, error)
	ChatsList(ctx context.Context, req *protos.ChatsListRequest) (*protos.ChatsListResponse, error)
}

type InterfaceStatisticRepo interface {
	GetStatistics(ctx context.Context, req *protos.StatisticsRequest) (*protos.StatisticsResponse, error)
}

type InterfacePromoCodeRepo interface {
	GetPromoCode(ctx context.Context, promoCode *entities.PromoCode) (*entities.PromoCode, error)
	PromoCodesList(ctx context.Context, page int32, limit int32, f entities.PromoCodeFilter) ([]entities.PromoCode, error)
	CreatePromoCode(ctx context.Context, promoCode *entities.PromoCode) error
	CountPromoCodes(ctx context.Context, f entities.PromoCodeFilter) (int64, error)
	UpdatePromoCode(ctx context.Context, promoCode *entities.PromoCode) error
}

type InterfaceReviewRepo interface {
	GetReview(ctx context.Context, review *entities.Review) (*entities.Review, error)
	ReviewsList(ctx context.Context, page int32, limit int32, f entities.ReviewFilter) ([]entities.Review, error)
	CountReviews(ctx context.Context, f entities.ReviewFilter) (int64, error)
}
