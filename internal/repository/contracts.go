package repository

import (
	"admin_history/internal/entities"
	protos "admin_history/pkg/proto/gen/go"
	"context"
)

type UserRepository interface {
	GetUser(ctx context.Context, user *entities.User) (*entities.User, error)
	UsersList(ctx context.Context, page int32, limit int32, f entities.UsersFilter) ([]entities.User, error)
	UpdateUser(ctx context.Context, user *entities.User) error
	CountUsers(ctx context.Context, f entities.UsersFilter) (int64, error)
}

type QuestionnaireRepository interface {
	GetQuestionnaire(ctx context.Context, questionnaire *entities.Questionnaire) (*entities.Questionnaire, error)
	GetQuestionnairesList(ctx context.Context, page int32, limit int32, f entities.QuestionnaireFilter) ([]entities.Questionnaire, error)
	UpdateQuestionnaire(ctx context.Context, req *entities.Questionnaire) error
	CountQuestionnaires(ctx context.Context, f entities.QuestionnaireFilter) (int64, error)
	SetQuestionnaireStatus(ctx context.Context, questionnaireID int64, status bool) error
}

type PhotoRepository interface {
	GetPhotosQuestionnaire(ctx context.Context, questionnaireID int64, typePhoto string) ([]entities.Photo, error)
	UploadPhoto(ctx context.Context, photo *entities.Photo) error
}

type VideoRepository interface {
	GetVideosQuestionnaire(ctx context.Context, questionnaireID int64, typeVideo string) ([]entities.Video, error)
	UploadVideo(ctx context.Context, video *entities.Video) error
}

type StatisticRepository interface {
	GetStatistics(ctx context.Context, req *protos.StatisticsRequest) (*protos.StatisticsResponse, error)
}

type PromoCodeRepository interface {
	GetPromoCode(ctx context.Context, promoCode *entities.PromoCode) (*entities.PromoCode, error)
	PromoCodesList(ctx context.Context, page int32, limit int32, f entities.PromoCodeFilter) ([]entities.PromoCode, error)
	CreatePromoCode(ctx context.Context, promoCode *entities.PromoCode) error
	CountPromoCodes(ctx context.Context, f entities.PromoCodeFilter) (int64, error)
	UpdatePromoCode(ctx context.Context, promoCode *entities.PromoCode) error
}

type ReviewRepository interface {
	GetReview(ctx context.Context, review *entities.Review) (*entities.Review, error)
	ReviewsList(ctx context.Context, page int32, limit int32, f entities.ReviewFilter) ([]entities.Review, error)
	CountReviews(ctx context.Context, f entities.ReviewFilter) (int64, error)
}
