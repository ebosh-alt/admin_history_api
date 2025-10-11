package usecase

import (
	"admin_history/config"
	"admin_history/internal/repository/postgres"
	"admin_history/internal/storage"
	"admin_history/internal/telegram"
	protos "admin_history/pkg/proto/gen/go"
	"context"

	"go.uber.org/zap"
)

type Usecase struct {
	cfg      *config.Config
	log      *zap.Logger
	Postgres postgres.InterfaceRepo
	ctx      context.Context
	st       *storage.FS
	tg       *telegram.Client
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

func NewUsecase(
	logger *zap.Logger,
	Postgres postgres.InterfaceRepo,
	cfg *config.Config,
	ctx context.Context,
	st *storage.FS,
	tg *telegram.Client,
) (*Usecase, error) {
	return &Usecase{
		cfg:      cfg,
		log:      logger,
		Postgres: Postgres,
		ctx:      ctx,
		st:       st,
		tg:       tg,
	}, nil
}

func (u *Usecase) OnStart(_ context.Context) error {
	return nil
}

func (u *Usecase) OnStop(_ context.Context) error {
	return nil
}

var _ InterfaceUsecase = (*Usecase)(nil)
