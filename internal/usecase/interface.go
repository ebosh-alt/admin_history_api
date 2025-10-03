package usecase

import (
	protos "admin_history/pkg/proto/gen/go"
	"context"
	"io"
)

type InterfaceUsecase interface {
	InterfaceUserUsecase
	InterfaceQuestionnaireUsecase
	InterfacePhotoUsecase
	InterfaceChatUsecase
	InterfaceStatisticUsecase
}

type InterfaceUserUsecase interface {
	GetUser(ctx context.Context, req *protos.UserRequest) (*protos.UserResponse, error)
	UsersList(ctx context.Context, req *protos.UsersListRequest) (*protos.UsersListResponse, error)
	UpdateUser(ctx context.Context, req *protos.UpdateUserRequest) (*protos.Status, error)
}

type InterfaceQuestionnaireUsecase interface {
	GetQuestionnairesList(ctx context.Context, req *protos.QuestionnairesListRequest) (*protos.QuestionnairesListResponse, error)
	GetQuestionnaire(ctx context.Context, req *protos.QuestionnaireRequest) (*protos.QuestionnaireResponse, error)
	UpdateQuestionnaire(ctx context.Context, req *protos.UpdateQuestionnaireRequest) (*protos.Status, error)
}

type InterfacePhotoUsecase interface {
	GetPhotosQuestionnaire(ctx context.Context, req *protos.PhotoRequest) (*protos.PhotoResponse, error)
	UploadPhoto(ctx context.Context, r io.Reader, filename string, photo *protos.Photo) (*protos.Status, error)
}

type InterfaceChatUsecase interface {
	GetChat(ctx context.Context, req *protos.ChatRequest) (*protos.ChatResponse, error)
	ChatsList(ctx context.Context, req *protos.ChatsListRequest) (*protos.ChatsListResponse, error)
}

type InterfaceStatisticUsecase interface {
	GetStatistics(ctx context.Context, req *protos.StatisticsRequest) (*protos.StatisticsResponse, error)
}
