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
}

type InterfaceUserRepo interface {
	GetUser(ctx context.Context, user *entities.User) (*entities.User, error)
	UsersList(ctx context.Context, page int32, limit int32, f entities.UsersFilter) ([]entities.User, error)
	UpdateUser(ctx context.Context, user *entities.User) error
}

type InterfaceQuestionnaireRepo interface {
	GetQuestionnaire(ctx context.Context, questionnaire *entities.Questionnaire) (*entities.Questionnaire, error)
	GetQuestionnairesList(ctx context.Context, page int32, limit int32, f entities.QuestionnaireFilter) ([]entities.Questionnaire, error)
	UpdateQuestionnaire(ctx context.Context, req *entities.Questionnaire) error
}

type InterfacePhotoRepo interface {
	GetPhotosQuestionnaire(ctx context.Context, questionnaireID int64, typePhoto string) ([]entities.Photo, error)
	UploadPhoto(ctx context.Context, photo *entities.Photo) error
}

type InterfaceChatRepo interface {
	GetChat(ctx context.Context, req *protos.ChatRequest) (*protos.ChatResponse, error)
	ChatsList(ctx context.Context, req *protos.ChatsListRequest) (*protos.ChatsListResponse, error)
}

type InterfaceStatisticRepo interface {
	GetStatistics(ctx context.Context, req *protos.StatisticsRequest) (*protos.StatisticsResponse, error)
}
