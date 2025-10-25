package usecase

import (
	"admin_history/config"
	"admin_history/internal/repository"
	"admin_history/internal/storage"
	protos "admin_history/pkg/proto/gen/go"
	"admin_history/pkg/telegram"
	"context"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Usecase struct {
	cfg *config.Config
	log *zap.Logger
	ctx context.Context
	st  *storage.FS
	tg  *telegram.Client

	users          repository.UserRepository
	questionnaires repository.QuestionnaireRepository
	photos         repository.PhotoRepository
	videos         repository.VideoRepository
	promoCodes     repository.PromoCodeRepository
	reviews        repository.ReviewRepository
}

func (u *Usecase) GetChat(ctx context.Context, req *protos.ChatRequest) (*protos.ChatResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (u *Usecase) ChatsList(ctx context.Context, req *protos.ChatsListRequest) (*protos.ChatsListResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (u *Usecase) GetStatistics(ctx context.Context, req *protos.StatisticsRequest) (*protos.StatisticsResponse, error) {
	//TODO implement me
	panic("implement me")
}

type RepositoryDeps struct {
	fx.In

	Users          repository.UserRepository
	Questionnaires repository.QuestionnaireRepository
	Photos         repository.PhotoRepository
	Videos         repository.VideoRepository
	PromoCodes     repository.PromoCodeRepository
	Reviews        repository.ReviewRepository
}

func NewUsecase(
	logger *zap.Logger,
	cfg *config.Config,
	ctx context.Context,
	st *storage.FS,
	tg *telegram.Client,
	repos RepositoryDeps,
) (*Usecase, error) {
	return &Usecase{
		cfg:            cfg,
		log:            logger,
		ctx:            ctx,
		st:             st,
		tg:             tg,
		users:          repos.Users,
		questionnaires: repos.Questionnaires,
		photos:         repos.Photos,
		videos:         repos.Videos,
		promoCodes:     repos.PromoCodes,
		reviews:        repos.Reviews,
	}, nil
}

func (u *Usecase) OnStart(_ context.Context) error {
	return nil
}

func (u *Usecase) OnStop(_ context.Context) error {
	return nil
}

var _ InterfaceUsecase = (*Usecase)(nil)
